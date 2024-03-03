package lsm

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/apache/arrow/go/v15/arrow/util"
	"github.com/dgraph-io/ristretto"
	"github.com/docker/go-units"
	"github.com/oklog/ulid/v2"
	"github.com/prometheus/client_golang/prometheus"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/events"
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

type Tree struct {
	ps     *PartStore
	index  index.Index
	mem    memory.Allocator
	merger *staples.Merger

	opts Options
	log  *slog.Logger

	nodes   []*RecordNode
	records []arrow.Record

	cache *ristretto.Cache
}

var (
	treeSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "vince",
		Subsystem: "lsm",
		Name:      "tree_size",
	})
	nodeSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "vince",
		Subsystem: "lsm",
		Name:      "node_size",
	})
	compactionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "vince",
		Subsystem: "lsm",
		Name:      "compaction_duration",
	})
	compactionCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "vince",
			Subsystem: "lsm",
			Name:      "num_compaction",
		},
	)
	nodesPerCompaction = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "vince",
			Subsystem: "lsm",
			Name:      "nodes_per_compaction",
		},
	)
)

func init() {
	prometheus.MustRegister(
		treeSize,
		nodeSize,
		compactionDuration,
		compactionCounter,
		nodesPerCompaction,
	)
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

func NewTree(mem memory.Allocator, indexer index.Index, opts ...Option) *Tree {
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
		opts:    o,
		nodes:   make([]*RecordNode, 0, 64),
		records: make([]arrow.Record, 0, 64),
		log: slog.Default().With(
			slog.String("component", "lsm-tree"),
		),
		cache: cache,
	}
}

func (lsm *Tree) Add(r arrow.Record) error {
	if r.NumRows() == 0 {
		return nil
	}

	idx, err := lsm.index.Index(r)
	if err != nil {
		return err
	}

	part := NewPart(r, idx)
	lsm.ps.Add(part)
	lsm.log.Debug("Added new part",
		"rows", r.NumRows(),
		"size", units.BytesSize(float64(part.size)))
	nodeSize.Observe(float64(part.Size()))
	treeSize.Observe(float64(lsm.Size()))
	return nil
}

func (lsm *Tree) Scan(ctx context.Context, start, end int64, fs *v1.Filters) (arrow.Record, error) {
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
	return lsm.ps.Scan(start, end, compiled, project), nil
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
			lsm.Compact(nil)
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

type RecordSource interface {
	Record(f func(arrow.Record) error) error
}

func (lsm *Tree) Restore(source RecordSource) error {
	lsm.ps.Reset()
	return source.Record(lsm.Add)
}

type CompactCallback func(r arrow.Record) bool

func (lsm *Tree) Compact(onCompact CompactCallback) {
	lsm.log.Debug("Start compaction")
	start := time.Now()
	stats := lsm.ps.Compact(func(r arrow.Record) {
		defer r.Release()
		if r.NumRows() == 0 {
			lsm.log.Debug("Skipping compaction, there is nothing in lsm tree")
			return
		}
		if onCompact != nil && !onCompact(r) {
			return
		}
		err := lsm.Add(r)
		if err != nil {
			lsm.log.Error("Failed adding compacted record to lsm", "tenant",
				"err", err)
			return
		}

	})
	lsm.log.Debug("Completed compaction", "elapsed", stats.Elapsed.String())
	compactionDuration.Observe(time.Since(start).Seconds())
	compactionCounter.Inc()
	nodesPerCompaction.Observe(float64(stats.CompactedNodesCount))
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
