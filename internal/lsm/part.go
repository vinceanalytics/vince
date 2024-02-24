package lsm

import (
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/vinceanalytics/vince/internal/closter/events"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/staples"
)

// PartStore is in memory storage of parts
type PartStore struct {
	size   atomic.Uint64
	tree   *RecordNode
	merger *staples.Merger
	nodes  []*RecordNode
}

func NewPartStore(mem memory.Allocator) *PartStore {
	return &PartStore{
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

func (p *PartStore) Scan(start, end int64, f func(index.Part)) {
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
			return true
		}
		return index.AcceptWith(
			int64(n.part.Min()),
			int64(n.part.Max()),
			start, end,
			func() {
				f(n.part)
			},
		)
	})
}

type CompactStats struct {
	OldSize             uint64
	CompactedNodesCount int
	Elapsed             time.Duration
}

func (p *PartStore) Compact() (r arrow.Record, stats CompactStats) {
	start := time.Now()
	defer func() {
		for _, r := range p.nodes {
			r.part.Release()
			r.part = nil
		}
		clear(p.nodes)
		stats.CompactedNodesCount = len(p.nodes)
		p.nodes = p.nodes[:0]
		stats.Elapsed = time.Since(start)
	}()
	p.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
			return true
		}
		p.merger.Add(n.part.Record())
		stats.OldSize += n.part.Size()
		p.nodes = append(p.nodes, n)
		return true
	})
	if stats.OldSize == 0 {
		r = p.merger.NewRecord()
		return
	}
	node := p.findNode(p.nodes[0])
	x := &RecordNode{}
	for !node.next.CompareAndSwap(p.nodes[0], x) {
		node = p.findNode(p.nodes[0])
	}
	p.size.Add(-stats.OldSize)
	r = p.merger.NewRecord()
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
