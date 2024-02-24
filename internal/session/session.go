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

type Tenants struct {
	cache    *ristretto.Cache
	sessions map[string]*Session
}

func (t *Tenants) Start(ctx context.Context) {
	for _, s := range t.sessions {
		s.Start(ctx)
	}
}

func (t *Tenants) Queue(ctx context.Context, resource string, req *v1.Event) {
	r, ok := t.sessions[resource]
	if !ok {
		return
	}
	r.Queue(ctx, req)
}

func (t *Tenants) Scan(ctx context.Context, resource string, start int64, end int64, fs *v1.Filters) (arrow.Record, error) {
	r, ok := t.sessions[resource]
	if !ok {
		return nil, ErrResourceNotFound
	}
	return r.Scan(ctx, start, end, fs)
}

func (t *Tenants) Close() {
	for _, s := range t.sessions {
		s.Close()
	}
}

func New(mem memory.Allocator, tenants *tenant.Tenants, storage db.Storage,
	indexer index.Index,
	primary index.Primary, opts ...lsm.Option) *Tenants {
	o := &Tenants{sessions: map[string]*Session{}}
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
	for _, v := range tenants.All() {
		o.sessions[v.Id] = newSession(
			mem, v.Id, cache, storage, indexer, primary, opts...,
		)
	}
	return o
}

type Session struct {
	list   *eventsv1.List
	build  *events.Builder
	events chan *v1.Event
	mu     sync.Mutex
	cache  *ristretto.Cache
	tree   *lsm.Tree
	log    *slog.Logger
}

func newSession(mem memory.Allocator, resource string, cache *ristretto.Cache, storage db.Storage,
	indexer index.Index,
	primary index.Primary, opts ...lsm.Option) *Session {

	return &Session{
		list:   events.List(),
		build:  events.New(mem),
		cache:  cache,
		events: make(chan *v1.Event, 4<<10),
		tree: lsm.NewTree(
			mem, resource, storage, indexer, primary, opts...,
		),
		log: slog.Default().With("component", "session"),
	}
}

func (s *Session) Queue(ctx context.Context, req *v1.Event) {
	s.events <- req
}

func (s *Session) doProcess(ctx context.Context) {
	s.log.Info("Starting events processing loop")
	defer func() {
		s.log.Info("Exiting events processing loop")
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-s.events:
			s.process(ctx, e)
		}
	}
}

func (s *Session) process(ctx context.Context, req *v1.Event) {
	e := events.Parse(ctx, req)
	if e == nil {
		return
	}
	events.Hit(e)
	if o, ok := s.cache.Get(e.Id); ok {
		cached := o.(*eventsv1.Data)
		s.mu.Lock()
		// cached can be accessed concurrently. Protect it together with build.
		events.Update(cached, e)
		s.list.Items = append(s.list.Items, e)
		s.mu.Unlock()
		return
	}
	s.mu.Lock()
	s.list.Items = append(s.list.Items, e)
	s.mu.Unlock()
	s.cache.SetWithTTL(e.Id, events.Clone(e), int64(proto.Size(e)), DefaultSession)
}

func (s *Session) Scan(ctx context.Context, start, end int64, fs *v1.Filters) (arrow.Record, error) {
	return s.tree.Scan(ctx, start, end, fs)
}

func (s *Session) Close() {
	s.tree.Compact(true)
}

func (s *Session) Flush() {
	s.mu.Lock()
	ls := s.list
	s.list = events.List()
	s.mu.Unlock()
	defer events.Put(ls)
	if len(ls.Items) == 0 {
		return
	}
	r := s.build.Write(ls)
	defer r.Release()
	s.log.Debug("Flushing sessions", "rows", r.NumRows())
	err := s.tree.Add(r)
	if err != nil {
		logger.Fail("Failed adding record to lsm", "err", err)
	}
}

func (s *Session) Start(ctx context.Context) {
	go s.tree.Start(ctx)
	go s.doFlush(ctx)
	go s.doProcess(ctx)
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

type sessionKey struct{}

func With(ctx context.Context, s *Tenants) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

func Get(ctx context.Context) *Tenants {
	return ctx.Value(sessionKey{}).(*Tenants)
}
