package store

import (
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

func (s *Sum) Update(ts time.Time, visitors, visits, events capnp.Float64List) {
	day := ts.YearDay()
	visitors.Set(day, visitors.At(day)+s.Visitors())
	visits.Set(day, visits.At(day)+s.Visits())
	events.Set(day, events.At(day)+s.Events())
}

func (c *Calendar) Update(ts time.Time, sum Sum) {
	visitors, _ := c.Visitors()
	visits, _ := c.Visits()
	events, _ := c.Events()
	sum.Update(ts, visitors, visits, events)
}

func ZeroCalendar(ts time.Time, sum Sum, f func([]byte) error) ([]byte, error) {
	var arena = capnp.MultiSegment(nil)
	msg, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, err
	}
	defer msg.Release()
	calendar, err := NewCalendar(seg)
	if err != nil {
		return nil, err
	}
	days := timex.DaysInAYear(ts)

	visits, err := capnp.NewFloat64List(seg, int32(days))
	if err != nil {
		return nil, err
	}

	visitors, err := capnp.NewFloat64List(seg, int32(days))
	if err != nil {
		return nil, err
	}

	events, err := capnp.NewFloat64List(seg, int32(days))
	if err != nil {
		return nil, err
	}
	sum.Update(ts, visitors, visits, events)
	err = calendar.SetVisitors(visitors)
	if err != nil {
		return nil, err
	}
	err = calendar.SetVisits(visits)
	if err != nil {
		return nil, err
	}
	err = calendar.SetEvents(events)
	if err != nil {
		return nil, err
	}
	return msg.MarshalPacked()
}

func (c *Calendar) SeriesVisitors(from, to time.Time) ([]float64, error) {
	ls, err := c.Visitors()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c *Calendar) SeriesVisits(from, to time.Time) ([]float64, error) {
	ls, err := c.Visits()
	if err != nil {
		return nil, err
	}
	return series(ls, from, to), nil
}

func (c *Calendar) SeriesEvents(from, to time.Time) ([]float64, error) {
	ls, err := c.Events()
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
