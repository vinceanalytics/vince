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
	eventsv1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cluster/events"
	"github.com/vinceanalytics/vince/internal/db"
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

var ErrResourceNotFound = errors.New("session: Resource not found")

func New(mem memory.Allocator, tenants *tenant.Tenants, storage db.Storage,
	indexer index.Index,
	primary index.Primary, opts ...lsm.Option) *Session {
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
		mem, cache, storage, indexer, primary, opts...,
	)
}

type Session struct {
	build *events.Multi
	mu    sync.Mutex
	cache *ristretto.Cache
	tree  *lsm.Tree
	log   *slog.Logger
}

func newSession(mem memory.Allocator, cache *ristretto.Cache, storage db.Storage,
	indexer index.Index,
	primary index.Primary, opts ...lsm.Option) *Session {

	return &Session{
		build: events.NewMulti(mem),
		cache: cache,
		tree: lsm.NewTree(
			mem, storage, indexer, primary, opts...,
		),
		log: slog.Default().With("component", "session"),
	}
}

func (s *Session) Persist() {
	s.Flush()
	s.tree.Compact(true)
}

func (s *Session) Append(e *v1.Data) {
	events.Hit(e)
	if o, ok := s.cache.Get(e.Id); ok {
		cached := o.(*eventsv1.Data)
		s.mu.Lock()
		// cached can be accessed concurrently. Protect it together with build.
		events.Update(cached, e)
		s.mu.Unlock()
		s.build.Append(e)
		return
	}
	s.build.Append(e)
	s.cache.SetWithTTL(e.Id, events.Clone(e), int64(proto.Size(e)), DefaultSession)
}

func (s *Session) Scan(ctx context.Context, tenantId string, start, end int64, fs *v1.Filters) (arrow.Record, error) {
	return s.tree.Scan(ctx, tenantId, start, end, fs)
}

func (s *Session) Close() {
	s.tree.Compact(true)
}

func (s *Session) Flush() {
	s.build.All(func(tenantId string, r arrow.Record) {
		// to avoid blocking ingestion add r in separate goroutine
		go s.append(tenantId, r)
	})
}
func (s *Session) append(tenantId string, r arrow.Record) {
	defer r.Release()
	err := s.tree.Add(tenantId, r)
	if err != nil {
		logger.Fail("Failed adding record to lsm", "tenant", tenantId, "err", err)
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
