package store

import (
	"errors"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/gernest/vince/timex"
)

//go:generate go run gen/main.go
const (
	// The cost of keeping a Calendar in memory
	CacheCost = 490704
)

func update(i int, v float64, f func() (capnp.Float64List, error)) error {
	if v == 0 {
		return nil
	}
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
	calendar, err := NewRootCalendar(seg)
	if err != nil {
		return nil, err
	}
	days := timex.CalendarHours(ts)
	cal := &calendar
	err = initFloats(days, seg,
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

func initFloats(n int, seg *capnp.Segment, fn ...func(capnp.Float64List) error) error {
	var errs []error
	for _, f := range fn {
		ls, err := capnp.NewFloat64List(seg, int32(n))
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
	var start, end int
	if from.IsZero() {
		start = 0
	} else {
		start = timex.HourIndex(from)
	}
	if to.IsZero() {
		end = f.Len()
	} else {
		end = timex.HourIndex(to)
	}
	if from.Equal(to) {
		start = 0
		end = f.Len()
	}
	diff := end - start
	o = make([]float64, diff)
	for i := 0; i < diff; i += 1 {
		o[i] = f.At(i + start)
	}
	return
}

func CalendarFromBytes(b []byte, f func(*Calendar) error) error {
	msg, err := capnp.UnmarshalPacked(b)
	if err != nil {
		return err
	}
	defer msg.Release()
	cal, err := ReadRootCalendar(msg)
	if err != nil {
		return err
	}
	return f(&cal)
}

type Sum struct {
	Visitors      uint32
	Views         uint32
	Events        uint32
	Visits        uint32
	BounceRate    uint32
	VisitDuration uint32
	ViewsPerVisit uint32
}

func (s *Sum) UpdateCalendar(ts time.Time, cal *Calendar) error {
	day := timex.HourIndex(ts)
	return errors.Join(
		update(day, float64(s.Visitors), cal.Visitors),
		update(day, float64(s.Views), cal.Views),
		update(day, float64(s.Events), cal.Events),
		update(day, float64(s.Visits), cal.Visits),
		update(day, float64(s.BounceRate), cal.BounceRate),
		update(day, float64(s.VisitDuration), cal.VisitDuration),
		update(day, float64(s.ViewsPerVisit), cal.ViewsPerVisit),
	)
}
