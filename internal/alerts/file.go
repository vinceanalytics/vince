package alerts

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/vinceanalytics/vince/internal/email"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/log"
)

type File struct {
	runtime *goja.Runtime
	mu      sync.Mutex
	calls   map[time.Duration]*Unit
}

type Unit struct {
	calls []goja.Callable
	file  *File
}

func (u *Unit) Call() {
	u.file.exec(u.calls)
}

func Create(ctx context.Context, js string) (*File, error) {
	s := &File{
		runtime: goja.New(),
		calls:   make(map[time.Duration]*Unit),
	}
	s.runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	query.Register(s.runtime)
	email.Register(ctx, s.runtime)
	s.runtime.Set("__schedule__", s.Schedule)
	s.runtime.Set("__query__", queryStats(ctx))
	_, err := s.runtime.RunString(js)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *File) Schedule(dur string, cb goja.Callable) {
	x, err := time.ParseDuration(dur)
	if err != nil {
		log.Get().Err(err).Str("duration", dur).Msg("invalid duration string")
		return
	}
	u, ok := s.calls[x]
	if !ok {
		u = &Unit{file: s}
		s.calls[x] = u
	}
	u.calls = append(u.calls, cb)
}

var ErrDomainNotFound = errors.New("Domain not found")

func queryStats(ctx context.Context) func(string, *query.Query) (*query.QueryResult, error) {
	return func(s string, q *query.Query) (*query.QueryResult, error) {
		site := models.SiteByDomain(ctx, s)
		if site == nil {
			return nil, ErrDomainNotFound
		}
		u := models.SiteOwner(ctx, site.ID)
		o := timeseries.Query(ctx, u.ID, site.ID, *q)
		return &o, nil
	}
}

func (s *File) exec(calls []goja.Callable) {
	s.mu.Lock()
	g := s.runtime.GlobalObject()
	for _, call := range calls {
		call(g)
	}
	s.mu.Unlock()
}

type Scheduler struct {
	units map[time.Duration][]*Unit
	g     sync.WaitGroup
	done  chan struct{}
}

func newScheduler() *Scheduler {
	return &Scheduler{
		units: make(map[time.Duration][]*Unit),
		done:  make(chan struct{}, 1),
	}
}

func (s *Scheduler) add(m map[time.Duration]*Unit) {
	for k, v := range m {
		s.units[k] = append(s.units[k], v)
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	for k, v := range s.units {
		go s.schedule(ctx, k, v)
	}
}

func (s *Scheduler) schedule(ctx context.Context, i time.Duration, calls []*Unit) {
	s.g.Add(1)
	defer s.g.Done()
	t := time.NewTicker(i)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case <-t.C:
			for _, v := range calls {
				v.Call()
			}
		}
	}
}

func (s *Scheduler) Close() error {
	s.done <- struct{}{}
	s.g.Wait()
	close(s.done)
	return nil
}
