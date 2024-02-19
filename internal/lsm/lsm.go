package lsm

import (
	"context"
	"encoding/binary"
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
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/docker/go-units"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/internal/camel"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/filters"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/staples"
)

type RecordPart struct {
	id     string
	record arrow.Record
	index.Full
	size uint64
}

var _ index.Part = (*RecordPart)(nil)

func (r *RecordPart) Record() arrow.Record {
	return r.record
}

func (r *RecordPart) Size() uint64 {
	return r.size
}

func (r *RecordPart) ID() string {
	return r.id
}

func (r *RecordPart) Release() {
	r.record.Release()
}

func NewPart(r arrow.Record, idx index.Full) *RecordPart {
	r.Retain()
	return &RecordPart{
		id:     ulid.Make().String(),
		record: r,
		Full:   idx,
		size:   uint64(util.TotalRecordSize(r)) + idx.Size(),
	}
}

type RecordNode = Node[index.Part]

type Tree[T any] struct {
	tree   *RecordNode
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

	nodes   []*RecordNode
	records []arrow.Record

	cache *ristretto.Cache
}

type Options struct {
	compactSize uint64
	ttl         time.Duration
}

const (
	compactSize = 16 << 20
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

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     int64(o.compactSize) * 2,
		BufferItems: 64,
		OnEvict: func(item *ristretto.Item) {
			if r, ok := item.Value.(arrow.Record); ok {
				r.Release()
			}
		},
		OnReject: func(item *ristretto.Item) {
			if r, ok := item.Value.(arrow.Record); ok {
				r.Release()
			}
		},
	})
	if err != nil {
		logger.Fail("Failed creating parts cache", "err", err)
	}
	return &Tree[T]{
		tree:     &RecordNode{},
		index:    indexer,
		mem:      mem,
		merger:   m,
		store:    db.NewStore(storage, mem, resource, o.ttl),
		primary:  primary,
		resource: resource,
		opts:     o,
		mapping:  mapping,
		schema:   schema,
		nodes:    make([]*RecordNode, 0, 64),
		records:  make([]arrow.Record, 0, 64),
		log: slog.Default().With(
			slog.String("component", "lsm-tree"),
			slog.String("resource", resource),
		),
		cache: cache,
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
	lsm.size.Add(part.size)
	lsm.tree.Prepend(part)
	lsm.log.Debug("Added new part", "size", units.BytesSize(float64(part.size)))
	return nil
}

func (lsm *Tree[T]) findNode(node *RecordNode) (list *RecordNode) {
	lsm.tree.Iterate(func(n *RecordNode) bool {
		if n.next.Load() == node {
			list = n
			return false
		}
		return true
	})
	return
}

func (lsm *Tree[T]) Scan(ctx context.Context, start, end int64, fs *v1.Filters) (arrow.Record, error) {
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

	lsm.ScanCold(ctx, start, end, compiled, project, func(r arrow.Record, ts *roaring.Bitmap) {
		tk(r, project, ts.ToArray())
	})

	lsm.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil {
			return true
		}
		return index.AcceptWith(int64(n.part.Min()), int64(n.part.Max()), start, end, func() {
			lsm.processPart(n.part, start, end, compiled, func(r arrow.Record, ts *roaring.Bitmap) {
				tk(r, project, ts.ToArray())
			})
		})
	})
	return tr.NewRecord(), nil
}

func (lsm *Tree[T]) ScanCold(ctx context.Context, start, end int64,
	compiled []*filters.CompiledFilter, columns []int, fn func(r arrow.Record, ts *roaring.Bitmap)) {
	granules := lsm.primary.FindGranules(lsm.resource, start, end)
	for _, granule := range granules {
		part := lsm.loadPart(ctx, granule, columns)
		if part != nil {
			lsm.processPart(part, start, end, compiled, fn)
		}
	}
}

func (lsm *Tree[T]) loadPart(ctx context.Context, id string, columns []int) index.Part {
	idx := lsm.loadIndex(ctx, id)
	if idx == nil {
		return nil
	}
	r := lsm.loadRecord(ctx, id, int64(idx.NumRows()), columns)
	if r == nil {
		return nil
	}
	return &db.Part{
		FileIndex: idx,
		Data:      r,
	}
}

func (lsm *Tree[T]) loadIndex(ctx context.Context, id string) *index.FileIndex {
	h := new(xxhash.Digest)
	h.WriteString(id)
	key := h.Sum64()
	v, ok := lsm.cache.Get(key)
	if ok {
		return v.(*index.FileIndex)
	}
	part, err := lsm.store.LoadIndex(ctx, id)
	if err != nil {
		lsm.log.Error("Failed loading granule index to memory", "id", id, "err", err)
		return nil
	}
	lsm.cache.Set(key, part, int64(part.Size()))
	return part
}

func (lsm *Tree[T]) loadRecord(ctx context.Context, id string, numRows int64, columns []int) arrow.Record {
	h := new(xxhash.Digest)
	h.WriteString(id)
	var a [8]byte
	binary.BigEndian.PutUint64(a[:], uint64(numRows))
	h.Write(a[:])
	for _, v := range columns {
		h.Write(binary.BigEndian.AppendUint32(a[:], uint32(v))[:4])
	}
	key := h.Sum64()
	v, ok := lsm.cache.Get(key)
	if ok {
		return v.(arrow.Record)
	}
	part, err := lsm.store.LoadRecord(ctx, id, numRows, columns)
	if err != nil {
		lsm.log.Error("Failed loading granule index to memory", "id", id, "err", err)
		return nil
	}
	lsm.cache.Set(key, part, util.TotalRecordSize(part))
	return part
}

func (lsm *Tree[T]) processPart(part index.Part, start, end int64,
	compiled []*filters.CompiledFilter, fn func(r arrow.Record, ts *roaring.Bitmap)) {
	r := part.Record()
	r.Retain()
	defer r.Release()
	ts := ScanTimestamp(r, lsm.mapping[v1.Filters_Timestamp.String()], start, end)
	part.Match(ts, compiled)
	if ts.IsEmpty() {
		return
	}
	fn(r, ts)
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

// Size returns the size in bytes of records+index in the lsm tree. This only
// accounts for active data.
//
// Cold data is still scanned by lsm tree but no account is about about its size.
func (lsm *Tree[T]) Size() uint64 {
	return lsm.Size()
}

func (lsm *Tree[T]) Compact(persist ...bool) {
	lsm.log.Debug("Start compaction")
	start := time.Now()
	defer func() {
		for _, r := range lsm.nodes {
			r.part.Release()
		}
		clear(lsm.nodes)
		clear(lsm.records)
		lsm.nodes = lsm.nodes[:0]
		lsm.records = lsm.records[:0]
	}()

	var oldSizes uint64
	lsm.tree.Iterate(func(n *RecordNode) bool {
		if n.part == nil || !n.part.CanIndex() {
			return true
		}
		lsm.nodes = append(lsm.nodes, n)
		lsm.records = append(lsm.records, n.part.Record())
		oldSizes += n.part.Size()
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
	x := &RecordNode{}
	for !node.next.CompareAndSwap(lsm.nodes[0], x) {
		node = lsm.findNode(lsm.nodes[0])
	}
	lsm.size.Add(-oldSizes)
	if oldSizes >= lsm.opts.compactSize || len(persist) > 0 {
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
	lsm.log.Debug("Saving compacted record to permanent storage")
	idx, err := lsm.index.Index(r)
	if err != nil {
		lsm.log.Error("Failed building index to persist", "err", err)
		return
	}
	granule, err := lsm.store.Save(r, idx)
	if err != nil {
		lsm.log.Error("Failed saving record", "err", err)
		return
	}
	lsm.primary.Add(lsm.resource, granule)
	lsm.log.Debug("Saved record to disc", "size", units.BytesSize(float64(granule.Size)))
	return
}

type Node[T any] struct {
	next atomic.Pointer[Node[T]]
	part T
}

func (n *Node[T]) Iterate(f func(*Node[T]) bool) {
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

func (n *Node[T]) Prepend(part T) *Node[T] {
	return n.prepend(&Node[T]{part: part})
}

func (n *Node[T]) prepend(node *Node[T]) *Node[T] {
	for {
		next := n.next.Load()
		node.next.Store(next)
		if n.next.CompareAndSwap(next, node) {
			return node
		}
	}
}
