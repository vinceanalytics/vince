package lsm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/arrow/util"
	"github.com/docker/go-units"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/camel"
	"github.com/vinceanalytics/vince/db"
	"github.com/vinceanalytics/vince/filters"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/index"
	"github.com/vinceanalytics/vince/staples"
)

type Part struct {
	ID     ulid.ULID
	Record arrow.Record
	Index  index.Full
	Size   uint64
	Min    int64
	Max    int64
}

func NewPart(r arrow.Record, idx index.Full) *Part {
	r.Retain()
	lo, hi := db.Timestamps(r)
	return &Part{
		ID:     ulid.Make(),
		Record: r,
		Index:  idx,
		Size:   uint64(util.TotalRecordSize(r)) + idx.Size(),
		Min:    lo,
		Max:    hi,
	}
}

type Tree[T any] struct {
	tree   *Node
	size   atomic.Uint64
	index  index.Index
	mem    memory.Allocator
	merger *staples.Merger
	store  *db.Store

	opts Options
	log  *slog.Logger

	primary  index.Primary
	resource string
	mapping  map[string]int
	schema   *arrow.Schema

	nodes   []*Node
	records []arrow.Record
}

type Options struct {
	compactSize uint64
	ttl         time.Duration
}

const (
	compactSize = 2 << 20
)

func DefaultLSMOptions() Options {
	return Options{
		compactSize: compactSize,
		ttl:         24 * 7 * time.Hour,
	}
}

type Option func(*Options)

func WithCompactSize(size uint64) Option {
	return func(l *Options) {
		l.compactSize = size
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(l *Options) {
		l.ttl = ttl
	}
}

func NewTree[T any](mem memory.Allocator, resource string, storage db.Storage, indexer index.Index, primary index.Primary, opts ...Option) *Tree[T] {
	schema := staples.Schema[T]()
	m := staples.NewMerger(mem, schema)
	mapping := make(map[string]int)
	for i, f := range schema.Fields() {
		mapping[f.Name] = i
	}
	o := DefaultLSMOptions()
	for _, f := range opts {
		f(&o)
	}
	return &Tree[T]{
		tree:     &Node{},
		index:    indexer,
		mem:      mem,
		merger:   m,
		store:    db.NewStore(storage, mem, resource, o.ttl),
		primary:  primary,
		resource: resource,
		opts:     o,
		mapping:  mapping,
		schema:   schema,
		nodes:    make([]*Node, 0, 64),
		records:  make([]arrow.Record, 0, 64),
		log: slog.Default().With(
			slog.String("component", "lsm-tree"),
			slog.String("resource", resource),
		),
	}
}

func (lsm *Tree[T]) Add(r arrow.Record) error {
	if r.NumRows() == 0 {
		return nil
	}

	idx, err := lsm.index.Index(r)
	if err != nil {
		return err
	}

	part := NewPart(r, idx)
	lsm.size.Add(part.Size)
	lsm.tree.Prepend(part)
	lsm.log.Debug("Added new part", "size", units.BytesSize(float64(part.Size)))
	return nil
}

func (lsm *Tree[T]) findNode(node *Node) (list *Node) {
	lsm.tree.Iterate(func(n *Node) bool {
		if n.next.Load() == node {
			list = n
			return false
		}
		return true
	})
	return
}

type ScanCallback func(context.Context, arrow.Record) error

func (lsm *Tree[T]) Scan(
	ctx context.Context,
	start, end int64,
	fs *v1.Filters,
) (arrow.Record, error) {
	ctx = compute.WithAllocator(ctx, lsm.mem)
	compiled, err := filters.CompileFilters(fs)
	if err != nil {
		lsm.log.Error("failed compiling scan filters", "err", err)
		return nil, err
	}
	if len(fs.Projection) == 0 {
		return nil, errors.New("missing projections")
	}
	project := make([]int, 0, len(fs.Projection))
	for _, name := range fs.Projection {
		col, ok := lsm.mapping[camel.Case(name.String())]
		if !ok {
			return nil, fmt.Errorf("column %s does not exist", name)
		}
		project = append(project, col)
	}
	fields := make([]arrow.Field, len(project))
	for i := range project {
		fields[i] = lsm.schema.Field(project[i])
	}
	schema := arrow.NewSchema(fields, nil)
	tr, tk := staples.NewTaker(lsm.mem, schema)
	defer tr.Release()

	lsm.tree.Iterate(func(n *Node) bool {
		if n.part == nil {
			return true
		}
		if n.part.Min <= end {
			if start <= n.part.Max {
				r := n.part.Record
				r.Retain()
				defer r.Release()
				ts := ScanTimestamp(r, lsm.mapping[v1.Filters_Timestamp.String()], start, end)
				n.part.Index.Match(ts, compiled)
				if ts.IsEmpty() {
					return true
				}
				tk(r, project, ts.ToArray())
				return true
			}
			return true
		}
		return false
	})
	return tr.NewRecord(), nil
}

func ScanTimestamp(r arrow.Record, timestampColumn int, start, end int64) *roaring.Bitmap {
	b := new(roaring.Bitmap)
	ls := r.Column(timestampColumn).(*array.Int64).Int64Values()
	from, _ := slices.BinarySearch(ls, start)
	to, _ := slices.BinarySearch(ls, end)
	for i := from; i < to; i++ {
		b.Add(uint32(i))
	}
	return b
}

func (lsm *Tree[T]) Start(ctx context.Context) {
	interval := 10 * time.Minute
	lsm.log.Info("Start compaction loop", "interval", interval.String(),
		"compactSize", units.BytesSize(float64(lsm.opts.compactSize)))
	tick := time.NewTicker(interval)
	defer func() {
		tick.Stop()
		lsm.log.Info("exiting compaction loop")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			lsm.Compact()
		}
	}

}

func (lsm *Tree[T]) Compact() {
	lsm.log.Debug("Start compaction")
	start := time.Now()
	defer func() {
		for _, r := range lsm.nodes {
			r.part.Record.Release()
		}
		clear(lsm.nodes)
		clear(lsm.records)
		lsm.nodes = lsm.nodes[:0]
		lsm.records = lsm.records[:0]
	}()

	var oldSizes uint64
	lsm.tree.Iterate(func(n *Node) bool {
		if n.part == nil {
			return true
		}
		lsm.nodes = append(lsm.nodes, n)
		lsm.records = append(lsm.records, n.part.Record)
		oldSizes += n.part.Size
		return true
	})
	if oldSizes == 0 {
		lsm.log.Debug("Skipping compaction, there is nothing in lsm tree")
		return
	}
	lsm.log.Debug("Compacting", "nodes", len(lsm.nodes), "size", oldSizes)
	r := lsm.merger.Merge(lsm.records...)
	defer r.Release()
	node := lsm.findNode(lsm.nodes[0])
	x := &Node{}
	for !node.next.CompareAndSwap(lsm.nodes[0], x) {
		node = lsm.findNode(lsm.nodes[0])
	}
	lsm.size.Add(-oldSizes)
	if oldSizes >= lsm.opts.compactSize {
		// Store in permanent storage
		lsm.log.Debug("Moving data to permanent storage")
		lsm.persist(r)
		return
	}
	err := lsm.Add(r)
	if err != nil {
		lsm.log.Error("Failed adding compacted record to lsm", "err", err)
		return
	}
	lsm.log.Debug("Completed compaction", "elapsed", time.Since(start).String())
}

func (lsm *Tree[T]) persist(r arrow.Record) {
	idx, err := lsm.index.Index(r)
	if err != nil {
		lsm.log.Error("Failed building index for record", "err", err)
		return
	}
	result, err := lsm.store.Save(r, idx)
	if err != nil {
		lsm.log.Error("Failed saving record to permanent storage", "err", err)
		return
	}
	lsm.log.Info("Saved indexed record to permanent storage",
		slog.String("id", result.Id),
		slog.Uint64("after_merge_size", result.Size),
		slog.Time("min_ts", time.Unix(0, int64(result.Min))),
		slog.Time("max_ts", time.Unix(0, int64(result.Max))),
	)
	lsm.primary.Add(lsm.resource, result)
	return
}

type Node struct {
	next atomic.Pointer[Node]
	part *Part
}

func (n *Node) Iterate(f func(*Node) bool) {
	if !(f(n)) {
		return
	}
	node := n.next.Load()
	for {
		if node == nil {
			return
		}
		if !f(node) {
			return
		}
		node = node.next.Load()
	}
}

func (n *Node) Prepend(part *Part) *Node {
	return n.prepend(&Node{part: part})
}

func (n *Node) prepend(node *Node) *Node {
	for {
		next := n.next.Load()
		node.next.Store(next)
		if n.next.CompareAndSwap(next, node) {
			return node
		}
	}
}
