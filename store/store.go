package store

import (
	"errors"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/gernest/vince/timex"
)

// Adds two sums
func (s *Sum) Add(o *Sum) {
	s.SetVisitors(s.Visitors() + o.Visitors())
	s.SetVisits(s.Visits() + o.Visits())
	s.SetEvents(s.Events() + o.Events())
}

func ZeroSum() Sum {
	_, seg, _ := capnp.NewMessage(capnp.MultiSegment(nil))
	sum, _ := NewSum(seg)
	sum.SetEvents(0)
	sum.SetVisitors(0)
	sum.SetVisits(0)
	return sum
}

func (s *Sum) SetValues(visitors, visits, views, events float64) {
	s.SetVisitors(visitors)
	s.SetVisits(visits)
	s.SetViews(views)
	s.SetEvents(events)
}

func (s *Sum) Reuse() *Sum {
	msg := s.Message()
	msg.Reset(capnp.MultiSegment(nil))
	s.SetEvents(0)
	s.SetVisitors(0)
	s.SetVisits(0)
	return s
}

func (s *Sum) UpdateCalendar(ts time.Time, cal *Calendar) error {
	day := ts.YearDay()
	return errors.Join(
		update(day, s.Visitors(), cal.Visitors),
		update(day, s.Views(), cal.Views),
		update(day, s.Events(), cal.Events),
		update(day, s.Visits(), cal.Visits),
		update(day, s.BounceRate(), cal.BounceRate),
		update(day, s.VisitDuration(), cal.VisitDuration),
		update(day, s.ViewsPerVisit(), cal.ViewsPerVisit),
	)
}

func update(i int, v float64, f func() (capnp.Float64List, error)) error {
	ls, err := f()
	if err != nil {
		return err
	}
	ls.Set(i, ls.At(i)+v)
	return nil
}

func ZeroCalendar(ts time.Time, sum *Sum) (*Calendar, error) {
	var arena = capnp.MultiSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}
	calendar, err := NewCalendar(seg)
	if err != nil {
		return nil, err
	}
	days := timex.DaysInAYear(ts)
	cal := &calendar
	err = initFloats(int32(days), seg,
		cal.SetVisitors,
		cal.SetViews,
		cal.SetEvents,
		cal.SetVisits,
		cal.SetBounceRate,
		cal.SetVisitDuration,
		cal.SetViewsPerVisit,
	)
	if err != nil {
		return nil, err
	}

	return cal, sum.UpdateCalendar(ts, cal)
}

func initFloats(n int32, seg *capnp.Segment, fn ...func(capnp.Float64List) error) error {
	var errs []error
	for _, f := range fn {
		ls, err := capnp.NewFloat64List(seg, n)
		if err != nil {
			return err
		}
		errs = append(errs, f(ls))
	}
	return errors.Join(errs...)
}

func (c *Calendar) SeriesVisitors(from, to time.Time) ([]float64, error) {
	ls, err := c.Visitors()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c Calendar) SeriesVisits(from, to time.Time) ([]float64, error) {
	ls, err := c.Visits()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c Calendar) SeriesEvents(from, to time.Time) ([]float64, error) {
	ls, err := c.Events()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c Calendar) SeriesViews(from, to time.Time) ([]float64, error) {
	ls, err := c.Views()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c Calendar) SeriesBounceRates(from, to time.Time) ([]float64, error) {
	ls, err := c.BounceRate()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c Calendar) SeriesVisitDuration(from, to time.Time) ([]float64, error) {
	ls, err := c.VisitDuration()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c Calendar) SeriesViewsPerVisit(from, to time.Time) ([]float64, error) {
	ls, err := c.ViewsPerVisit()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func series(f capnp.Float64List, from, to time.Time) (o []float64) {
	if from.Year() != to.Year() || to.Before(from) {
		return
	}
	start := from.YearDay()
	end := to.YearDay()

	o = make([]float64, end-start)
	for i := 0; i < end-start; i += 1 {
		o[i] = f.At(i + start)
	}
	return
}

func CalendarFromBytes(b []byte) (Calendar, error) {
	msg, err := capnp.UnmarshalPacked(b)
	if err != nil {
		return Calendar{}, err
	}
	return ReadRootCalendar(msg)
}

func SumFromBytes(b []byte) (Sum, error) {
	msg, err := capnp.UnmarshalPacked(b)
	if err != nil {
		return Sum{}, err
	}
	return ReadRootSum(msg)
}
