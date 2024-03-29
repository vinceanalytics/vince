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
	nodes  []*RecordNode
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

func (p *PartStore) Scan(start, end int64,
	compiled []*filters.CompiledFilter,
	projected []string) arrow.Record {
	b, take := staples.NewTaker(p.mem, projected)
	defer b.Release()
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
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

func (p *PartStore) Reset() {
	next := p.tree.next.Load()
	p.tree.next.Store(nil)
	if next != nil {
		// release all existing parts
		next.Iterate(func(n *RecordNode) bool {
			if n.part != nil {
				n.part.Release()
			}
			return true
		})
	}

}
func (p *PartStore) Compact(f func(r arrow.Record)) (stats CompactStats) {

	start := time.Now()
	defer func() {
		for _, r := range p.nodes {
			r.part.Release()
			r.part = nil
		}
		stats.CompactedNodesCount = len(p.nodes)

		clear(p.nodes)
		p.nodes = p.nodes[:0]

		stats.Elapsed = time.Since(start)
	}()
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
			return true
		}
		stats.OldSize += n.part.Size()
		p.merger.Merge(n.part.Record())
		p.nodes = append(p.nodes, n)
		return true
	})
	if len(p.nodes) == 0 {
		return
	}

	r := p.merger.NewRecord(arrow.Metadata{})
	f(r)
	first := p.nodes[0]
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
