package session

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/apache/arrow/go/v15/arrow/memory"
	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/staples/staples/db"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/index"
	"github.com/vinceanalytics/staples/staples/logger"
	"github.com/vinceanalytics/staples/staples/lsm"
	"github.com/vinceanalytics/staples/staples/staples"
)

const (
	DefaultSession = 15 * time.Minute
	// To make sure we always have fresh data for current visitors
	DefaultFlushInterval = 5 * time.Minute
)

type Session struct {
	build *staples.Arrow[staples.Event]
	mu    sync.Mutex
	cache *ristretto.Cache
	tree  *lsm.Tree[staples.Event]
	log   *slog.Logger
}

func New(mem memory.Allocator, resource string, storage db.Storage,
	indexer index.Index,
	primary index.Primary, opts ...lsm.Option) *Session {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     2 << 20,
		BufferItems: 64,
	})
	if err != nil {
		logger.Fail("Failed initializing cache", "err", err)
	}
	return &Session{
		build: staples.NewArrow[staples.Event](mem),
		cache: cache,
		tree: lsm.NewTree[staples.Event](
			mem, resource, storage, indexer, primary, opts...,
		),
		log: slog.Default().With("component", "session"),
	}
}

func (s *Session) Queue(ctx context.Context, req *v1.Event) {
	e := staples.Parse(ctx, req)
	if e == nil {
		return
	}
	e.Hit()
	if o, ok := s.cache.Get(e.ID); ok {
		cached := o.(*staples.Event)
		cached.Update(e)
		defer e.Release()
	} else {
		s.cache.SetWithTTL(e.ID, e, 1, DefaultSession)
	}
	s.mu.Lock()
	s.build.Append(e)
	s.mu.Unlock()
}

func (s *Session) Flush() {
	s.mu.Lock()
	r := s.build.NewRecord()
	s.mu.Unlock()
	if r.NumRows() == 0 {
		r.Release()
		return
	}
	err := s.tree.Add(r)
	if err != nil {
		logger.Fail("Failed adding record to lsm", "err", err)
	}
}

func (s *Session) Start(ctx context.Context) {
	go s.doFlush(ctx)
}

func (s *Session) doFlush(ctx context.Context) {

	s.log.Info("Starting session flushing loop", slog.Duration("interval", DefaultFlushInterval))
	tk := time.NewTicker(DefaultFlushInterval)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			s.Flush()
		}
	}
}

type sessionKey struct{}

func With(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

func Get(ctx context.Context) *Session {
	return ctx.Value(sessionKey{}).(*Session)
}
