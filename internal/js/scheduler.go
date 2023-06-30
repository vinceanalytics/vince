package js

import (
	"context"
	"sync"
	"time"
)

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
