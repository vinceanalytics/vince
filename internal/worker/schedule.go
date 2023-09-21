package worker

import (
	"context"
	"log/slog"
	"sort"
	"time"

	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/g"
	"golang.org/x/sync/errgroup"
)

type JobScheduler struct {
	add    chan *jobEntry
	remove chan string
	log    *slog.Logger
	e      []*jobEntry
	g      *errgroup.Group
}

func (j *JobScheduler) Close() error {
	close(j.add)
	close(j.remove)
	j.log.Info("stopped scheduler")
	return nil
}

func (s *JobScheduler) Schedule(id string, schedule Schedule, job Job) {
	s.add <- &jobEntry{
		id:       id,
		job:      job,
		schedule: schedule,
	}
}

type schedulerKey struct{}

func OpenScheduler(ctx context.Context) (context.Context, *JobScheduler) {
	s := &JobScheduler{
		add:    make(chan *jobEntry),
		remove: make(chan string),
		log:    slog.Default().With("component", "scheduler"),
		g:      g.Get(ctx),
	}
	g.Get(ctx).Go(func() error {
		s.run(ctx)
		return nil
	})
	return context.WithValue(ctx, schedulerKey{}, s), s
}

func GetScheduler(ctx context.Context) *JobScheduler {
	return ctx.Value(schedulerKey{}).(*JobScheduler)
}

func (s *JobScheduler) run(ctx context.Context) {
	s.log.Info("starting scheduler")
	now := core.Now(ctx)
	for _, x := range s.e {
		x.next = x.schedule.next(now)
	}
	sleepFor := 100000 * time.Hour
	sleep := time.NewTimer(sleepFor)
	defer sleep.Stop()
	for {
		sort.Sort(byTime(s.e))
		if len(s.e) == 0 || s.e[0].next.IsZero() {
			sleep.Reset(sleepFor)
		} else {
			sleep.Reset(s.e[0].next.Sub(now))
		}
		for {
			select {
			case <-ctx.Done():
				return
			case now = <-sleep.C:
				s.log.Info("waking up")
				for _, x := range s.e {
					if x.next.After(now) || x.next.IsZero() {
						break
					}
					s.g.Go(runJob(ctx, x.job))
					x.prev = x.next
					x.next = x.schedule.next(now)
					s.log.Info("running job", "job_id", x.id)
				}
			case x := <-s.add:
				sleep.Reset(sleepFor)
				now = core.Now(ctx)
				x.next = x.schedule.next(now)
				s.e = append(s.e, x)
				s.log.Info("added new job", "job_id", x.id)
			case id := <-s.remove:
				o := make([]*jobEntry, 0, len(s.e))
				for _, v := range s.e {
					if v.id != id {
						o = append(o, v)
					}
				}
				s.e = o
				s.log.Info("removed  job", "job_id", id)
			}
			break
		}
	}
}

func runJob(ctx context.Context, j Job) func() error {
	return func() error {
		j.Run(ctx)
		return nil
	}
}

type Job interface {
	Run(context.Context)
}

type JobFunc func(context.Context)

var _ Job = (*JobFunc)(nil)

func (f JobFunc) Run(ctx context.Context) {
	f(ctx)
}

type jobEntry struct {
	id         string
	schedule   Schedule
	prev, next time.Time
	job        Job
}

type byTime []*jobEntry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].next.IsZero() {
		return false
	}
	if s[j].next.IsZero() {
		return true
	}
	return s[i].next.Before(s[j].next)
}

type Schedule interface {
	next(time.Time) time.Time
}

// Daily schedules the job to run at midnight
type Daily struct{}

var _ Schedule = (*Daily)(nil)

func (Daily) next(ts time.Time) time.Time {
	return time.Date(ts.Year(), ts.Month(), ts.Day(), 24, 0, 0, 0, ts.Location())
}

type Every time.Duration

var _ Schedule = (*Every)(nil)

func (e Every) next(ts time.Time) time.Time {
	return ts.Add(time.Duration(e))
}
