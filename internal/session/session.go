package session

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/dgraph-io/ristretto"
	"github.com/prometheus/client_golang/prometheus"
	eventsv1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/index"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/lsm"
	"github.com/vinceanalytics/vince/internal/tenant"
	"google.golang.org/protobuf/proto"
)

const (
	DefaultSession = 30 * time.Minute
	// To make sure we always have fresh data for current visitors
	DefaultFlushInterval = time.Minute
)

var (
	eventsCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vince",
		Subsystem: "session",
		Name:      "events_total",
	})
	sessionCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "vince",
		Subsystem: "session",
		Name:      "sessions_total",
	})
)

func init() {
	prometheus.MustRegister(eventsCount, sessionCount)
}

var ErrResourceNotFound = errors.New("session: Resource not found")

func New(mem memory.Allocator, tenants *tenant.Tenants,
	indexer index.Index, opts ...lsm.Option) *Session {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     100 << 20, // 100MiB
		BufferItems: 64,
		OnEvict: func(item *ristretto.Item) {
			e := item.Value.(*eventsv1.Data)
			events.PutOne(e)
			item.Value = nil
		},
		OnReject: func(item *ristretto.Item) {
			e := item.Value.(*eventsv1.Data)
			events.PutOne(e)
			item.Value = nil
		},
	})
	if err != nil {
		logger.Fail("Failed initializing cache", "err", err)
	}
	return newSession(
		mem, cache, indexer, opts...,
	)
}

type Session struct {
	build *events.Builder
	mu    sync.Mutex
	cache *ristretto.Cache
	tree  *lsm.Tree
	log   *slog.Logger
}

func newSession(mem memory.Allocator, cache *ristretto.Cache,
	indexer index.Index, opts ...lsm.Option) *Session {

	return &Session{
		build: events.New(mem),
		cache: cache,
		tree: lsm.NewTree(
			mem, indexer, opts...,
		),
		log: slog.Default().With("component", "session"),
	}
}

func (s *Session) Restore(source lsm.RecordSource) error {
	return s.tree.Restore(source)
}

func (s *Session) Persist(onCompact lsm.CompactCallback) {
	s.Flush()
	s.tree.Compact(onCompact)
}

func (s *Session) Append(e *v1.Data) {
	eventsCount.Inc()
	events.Hit(e)
	if o, ok := s.cache.Get(e.Id); ok {
		cached := o.(*eventsv1.Data)
		s.mu.Lock()
		// cached can be accessed concurrently. Protect it together with build.
		events.Update(cached, e)
		s.mu.Unlock()
		s.build.WriteData(e)
		return
	}
	s.build.WriteData(e)
	s.cache.SetWithTTL(e.Id, events.Clone(e), int64(proto.Size(e)), DefaultSession)
	sessionCount.Inc()
}

func (s *Session) Scan(ctx context.Context, start, end int64, fs *v1.Filters) (arrow.Record, error) {
	return s.tree.Scan(ctx, start, end, fs)
}

func (s *Session) Close() {
	s.tree.Compact(nil)
}

func (s *Session) Flush() {
	r := s.build.NewRecord()
	defer r.Release()
	err := s.tree.Add(r)
	if err != nil {
		logger.Fail("Failed adding record to lsm", "err", err)
	}
}

func (s *Session) Start(ctx context.Context) {
	go s.tree.Start(ctx)
	go s.doFlush(ctx)
}

func (s *Session) doFlush(ctx context.Context) {
	s.log.Info("Starting session flushing loop", "interval", DefaultFlushInterval.String())

	tk := time.NewTicker(DefaultFlushInterval)
	defer func() {
		tk.Stop()
		s.log.Info("Exiting flushing loop")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			s.Flush()
		}
	}
}
