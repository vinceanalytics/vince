package js

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

type Runnable interface {
	Interval() time.Duration
	Run()
}

type UnitList struct {
	i time.Duration
	u []*Unit
}

var _ Runnable = (*UnitList)(nil)

func (u *UnitList) Interval() time.Duration {
	return u.i
}
func (u *UnitList) Run() {
	for _, c := range u.u {
		c.Run()
	}
}

type Unit struct {
	i     time.Duration
	calls []goja.Callable
	file  *File
}

var _ Runnable = (*Unit)(nil)

func (u *Unit) Interval() time.Duration {
	return u.i
}

func (u *Unit) Run() {
	u.file.exec(u.calls)
}

func create(ctx context.Context) *File {
	s := &File{
		runtime: goja.New(),
		calls:   make(map[time.Duration]*Unit),
	}
	s.runtime.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	query.Register(s.runtime)
	email.Register(ctx, s.runtime)
	s.runtime.Set("__schedule__", s.Schedule)
	s.runtime.Set("__query__", queryStats(ctx))
	return s
}

func (s *File) Schedule(dur string, cb goja.Callable) {
	x, err := time.ParseDuration(dur)
	if err != nil {
		log.Get().Err(err).Str("duration", dur).Msg("invalid duration string")
		return
	}
	u, ok := s.calls[x]
	if !ok {
		u = &Unit{file: s, i: x}
		s.calls[x] = u
	}
	u.calls = append(u.calls, cb)
}

var ErrDomainNotFound = errors.New("domain not found")

func queryStats(ctx context.Context) func(string, *query.Query) (*query.QueryResult, error) {
	return func(s string, q *query.Query) (*query.QueryResult, error) {
		site := models.SiteByDomain(ctx, s)
		if site == nil {
			return nil, ErrDomainNotFound
		}
		o := timeseries.Query(ctx, site.UserID, site.ID, *q)
		return &o, nil
	}
}

func (f *File) Units() (o []Runnable) {
	for _, v := range f.calls {
		o = append(o, v)
	}
	return
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
	units map[time.Duration][]Runnable
	g     sync.WaitGroup
	done  chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		units: make(map[time.Duration][]Runnable),
		done:  make(chan struct{}, 1),
	}
}

func (s *Scheduler) Add(r Runnable) {
	k := r.Interval()
	s.units[k] = append(s.units[k], r)
}

func (s *Scheduler) Run(ctx context.Context) {
	for k, v := range s.units {
		s.g.Add(1)
		go s.schedule(ctx, k, v)
	}
	s.g.Wait()
}

func (s *Scheduler) schedule(ctx context.Context, i time.Duration, calls []Runnable) {
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
				v.Run()
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
