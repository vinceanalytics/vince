package lsm

import (
	"context"
	"encoding/binary"
	"errors"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/RoaringBitmap/roaring"
	"github.com/VictoriaMetrics/metrics"
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/arrow/util"
	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/ristretto"
	"github.com/docker/go-units"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/columns"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/filters"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/staples"
)

type RecordPart struct {
	tenant string
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

func NewPart(tenantId string, r arrow.Record, idx index.Full) *RecordPart {
	r.Retain()
	return &RecordPart{
		tenant: tenantId,
		id:     ulid.Make().String(),
		record: r,
		Full:   idx,
		size:   uint64(util.TotalRecordSize(r)) + idx.Size(),
	}
}

func (r *RecordPart) Tenant() string {
	return r.tenant
}

type RecordNode = Node[index.Part]

type Tree struct {
	ps     *PartStore
	index  index.Index
	mem    memory.Allocator
	merger *staples.Merger
	store  *db.Store

	opts Options
	log  *slog.Logger

	primary index.Primary

	nodes   []*RecordNode
	records []arrow.Record

	cache *ristretto.Cache
}

var (
	treeSize           = metrics.NewHistogram("vnc_lsm_tree_size{")
	nodeSize           = metrics.NewHistogram("vnc_lsm_node_size{")
	compactionDuration = metrics.NewHistogram("vnc_lsm_compaction_duration_seconds{")
	compactionCounter  = metrics.NewCounter("vnc_lsm_compaction{")
	nodesPerCompaction = metrics.NewHistogram("vnc_lsm_nodes_per_compaction{")
)

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

func NewTree(mem memory.Allocator, storage db.Storage, indexer index.Index, primary index.Primary, opts ...Option) *Tree {
	m := staples.NewMerger(mem, events.Schema)
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
	return &Tree{
		ps:      NewPartStore(mem),
		index:   indexer,
		mem:     mem,
		merger:  m,
		store:   db.NewStore(storage, mem, o.ttl),
		primary: primary,
		opts:    o,
		nodes:   make([]*RecordNode, 0, 64),
		records: make([]arrow.Record, 0, 64),
		log: slog.Default().With(
			slog.String("component", "lsm-tree"),
		),
		cache: cache,
	}
}

func (lsm *Tree) Add(tenantId string, r arrow.Record) error {
	if r.NumRows() == 0 {
		return nil
	}

	idx, err := lsm.index.Index(r)
	if err != nil {
		return err
	}

	part := NewPart(tenantId, r, idx)
	lsm.ps.Add(part)
	lsm.log.Debug("Added new part", "tenant", tenantId, "size", units.BytesSize(float64(part.size)))
	nodeSize.Update(float64(part.Size()))
	treeSize.Update(float64(lsm.Size()))
	return nil
}

func (lsm *Tree) Scan(ctx context.Context, tenantId string, start, end int64, fs *v1.Filters) (arrow.Record, error) {
	ctx = compute.WithAllocator(ctx, lsm.mem)
	compiled, err := filters.CompileFilters(fs)
	if err != nil {
		lsm.log.Error("failed compiling scan filters", "err", err)
		return nil, err
	}
	if len(fs.Projection) == 0 {
		return nil, errors.New("missing projections")
	}
	project := make([]string, 0, len(fs.Projection))
	for _, name := range fs.Projection {
		project = append(project, name.String())
	}
	return lsm.ps.Scan(tenantId, start, end, compiled, project), nil
}

func (lsm *Tree) ScanCold(ctx context.Context, tenant string, start, end int64,
	compiled []*filters.CompiledFilter, columns []int, fn func(r arrow.Record, ts *roaring.Bitmap)) {
	granules := lsm.primary.FindGranules(tenant, start, end)
	for _, granule := range granules {
		part := lsm.loadPart(ctx, tenant, granule, columns)
		if part != nil {
			lsm.processPart(part, start, end, compiled, fn)
		}
	}
}

func (lsm *Tree) loadPart(ctx context.Context, tenant, id string, columns []int) index.Part {
	idx := lsm.loadIndex(ctx, tenant, id)
	if idx == nil {
		return nil
	}
	r := lsm.loadRecord(ctx, tenant, id, int64(idx.NumRows()), columns)
	if r == nil {
		return nil
	}
	return &db.Part{
		FileIndex: idx,
		Data:      r,
	}
}

func (lsm *Tree) loadIndex(ctx context.Context, tenant, id string) *index.FileIndex {
	h := new(xxhash.Digest)
	h.WriteString(id)
	key := h.Sum64()
	v, ok := lsm.cache.Get(key)
	if ok {
		return v.(*index.FileIndex)
	}
	part, err := lsm.store.LoadIndex(ctx, tenant, id)
	if err != nil {
		lsm.log.Error("Failed loading granule index to memory", "id", id, "err", err)
		return nil
	}
	lsm.cache.Set(key, part, int64(part.Size()))
	return part
}

func (lsm *Tree) loadRecord(ctx context.Context, tenant, id string, numRows int64, columns []int) arrow.Record {
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
	part, err := lsm.store.LoadRecord(ctx, tenant, id, numRows, columns)
	if err != nil {
		lsm.log.Error("Failed loading granule index to memory", "id", id, "err", err)
		return nil
	}
	lsm.cache.Set(key, part, util.TotalRecordSize(part))
	return part
}

func (lsm *Tree) processPart(part index.Part, start, end int64,
	compiled []*filters.CompiledFilter, fn func(r arrow.Record, ts *roaring.Bitmap)) {
	r := part.Record()
	r.Retain()
	defer r.Release()
	ts := ScanTimestamp(r, events.Mapping[columns.Timestamp], start, end)
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

func (lsm *Tree) Start(ctx context.Context) {
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
func (lsm *Tree) Size() uint64 {
	return lsm.ps.Size()
}

func (lsm *Tree) Compact(persist ...bool) {
	lsm.log.Debug("Start compaction")
	start := time.Now()
	var save bool
	if len(persist) > 0 {
		save = persist[0]
	}
	stats := lsm.ps.Compact(func(tenant string, r arrow.Record) {
		defer r.Release()
		if r.NumRows() == 0 {
			lsm.log.Debug("Skipping compaction, there is nothing in lsm tree", "tenant", tenant)
			return
		}
		if save {
			lsm.persist(tenant, r)
		} else {
			err := lsm.Add(tenant, r)
			if err != nil {
				lsm.log.Error("Failed adding compacted record to lsm", "tenant", tenant, "err", err)
				return
			}
		}
	})
	lsm.log.Debug("Completed compaction", "elapsed", stats.Elapsed.String())
	compactionDuration.UpdateDuration(start)
	compactionCounter.Inc()
	nodesPerCompaction.Update(float64(stats.CompactedNodesCount))
}

func (lsm *Tree) persist(tenant string, r arrow.Record) {
	lsm.log.Debug("Saving compacted record to permanent storage", "tenant", tenant)
	idx, err := lsm.index.Index(r)
	if err != nil {
		lsm.log.Error("Failed building index to persist", "tenant", tenant, "err", err)
		return
	}
	granule, err := lsm.store.Save(tenant, r, idx)
	if err != nil {
		lsm.log.Error("Failed saving record", "tenant", tenant, "err", err)
		return
	}
	lsm.primary.Add(tenant, granule)
	lsm.log.Debug("Saved record to disc", "tenant", tenant, "size", units.BytesSize(float64(granule.Size)))
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
