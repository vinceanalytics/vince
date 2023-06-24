package js

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

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

type Scheduler struct {
	units map[time.Duration][]*Alert
	g     sync.WaitGroup
	done  chan struct{}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		units: make(map[time.Duration][]*Alert),
		done:  make(chan struct{}, 1),
	}
}

func (s *Scheduler) Add(a *Alert) {
	s.units[a.Interval] = append(s.units[a.Interval], a)
}

func (s *Scheduler) Run(ctx context.Context) {
	for k, v := range s.units {
		s.g.Add(1)
		go s.schedule(ctx, k, v)
	}
	s.g.Wait()
}

func (s *Scheduler) schedule(ctx context.Context, i time.Duration, calls []*Alert) {
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
				v.Run(ctx)
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
