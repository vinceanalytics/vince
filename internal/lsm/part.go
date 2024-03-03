package lsm

import (
	"slices"
	"sync/atomic"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/filters"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/staples"
)

// PartStore is in memory storage of parts
type PartStore struct {
	mem    memory.Allocator
	size   atomic.Uint64
	tree   *RecordNode
	merger *staples.Merger
}

func NewPartStore(mem memory.Allocator) *PartStore {
	return &PartStore{
		mem:    mem,
		tree:   new(RecordNode),
		merger: staples.NewMerger(mem, events.Schema),
	}
}

func (p *PartStore) Release() {
	p.merger.Release()
}

func (p *PartStore) Size() uint64 {
	return p.size.Load()
}

func (p *PartStore) Add(r *RecordPart) {
	p.size.Add(r.Size())
	p.tree.Prepend(r)
}

func (p *PartStore) Scan(tenant string, start, end int64,
	compiled []*filters.CompiledFilter,
	projected []string) arrow.Record {
	b, take := staples.NewTaker(p.mem, projected)
	defer b.Release()
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
			return true
		}
		if n.part.Tenant() != tenant {
			return true
		}
		return index.AcceptWith(
			int64(n.part.Min()),
			int64(n.part.Max()),
			start, end,
			func() {
				r := n.part.Record()
				r.Retain()
				defer r.Release()
				tsCol := r.Column(0).(*array.Int64)
				ts := scanRange(tsCol.Int64Values(), start, end)
				n.part.Match(ts, compiled)
				if ts.IsEmpty() {
					return
				}
				take(r, ts.ToArray())
			},
		)
	})
	r := b.NewRecord()
	return r
}

func scanRange(ls []int64, start, end int64) *roaring.Bitmap {
	b := new(roaring.Bitmap)
	from, _ := slices.BinarySearch(ls, start)
	to, _ := slices.BinarySearch(ls, end)
	for i := from; i < to; i++ {
		b.Add(uint32(i))
	}
	return b
}

type CompactStats struct {
	OldSize             uint64
	CompactedNodesCount int
	Elapsed             time.Duration
}

func (p *PartStore) Compact(f func(tenant string, r arrow.Record)) (stats CompactStats) {
	ns := make(map[string][]*RecordNode)
	start := time.Now()
	defer func() {
		for _, nodes := range ns {
			for _, r := range nodes {
				r.part.Release()
				r.part = nil
			}
			stats.CompactedNodesCount += len(nodes)
			clear(nodes)
		}
		stats.Elapsed = time.Since(start)
	}()
	var first *RecordNode
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
			return true
		}
		if first == nil {
			first = n
		}
		ns[n.part.Tenant()] = append(ns[n.part.Tenant()], n)
		return true
	})
	if len(ns) == 0 {
		return
	}
	for tenant, n := range ns {
		stats = CompactStats{
			CompactedNodesCount: len(n),
		}
		start := time.Now()
		for _, v := range n {
			stats.OldSize += v.part.Size()
			p.merger.Merge(v.part.Record())
		}
		stats.Elapsed = time.Since(start)
		r := p.merger.NewRecord(arrow.MetadataFrom(map[string]string{
			"tenant_id": tenant,
		}))
		f(tenant, r)
	}
	node := p.findNode(first)
	x := &RecordNode{}
	for !node.next.CompareAndSwap(first, x) {
		node = p.findNode(first)
	}
	p.size.Add(-stats.OldSize)
	return
}

func (p *PartStore) findNode(node *RecordNode) (list *RecordNode) {
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.next.Load() == node {
			list = n
			return false
		}
		return true
	})
	return
}
