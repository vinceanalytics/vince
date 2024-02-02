package session

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/vinceanalytics/staples"
	v1 "github.com/vinceanalytics/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/logger"
	"github.com/vinceanalytics/staples/lsm"
)

const DefaultSession = 15 * time.Minute

type Session struct {
	build *staples.Arrow[staples.Event]
	mu    sync.Mutex
	cache *ristretto.Cache
	tree  *lsm.Tree[staples.Event]
	log   *slog.Logger
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
		return
	}
	err := s.tree.Add(r)
	if err != nil {
		logger.Fail("Failed adding record to lsm", "err", err)
	}
}

type sessionKey struct{}

func With(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

func Get(ctx context.Context) *Session {
	return ctx.Value(sessionKey{}).(*Session)
}
