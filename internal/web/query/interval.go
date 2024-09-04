package query

import (
	"time"

	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

type Interval byte

const (
	Minute Interval = 1 + iota
	Hour
	Day
	Week
	Month
)

func (i Interval) Format() string {
	switch i {
	case Minute, Hour, Day:
		return time.DateTime
	default:
		return time.DateOnly
	}
}

func (i Interval) Reduce(ts int64) time.Time {
	t := time.UnixMilli(ts).UTC()
	switch i {
	case Minute:
		return t.Truncate(time.Minute)
	case Hour:
		return t.Truncate(time.Hour)
	case Day:
		yy, mm, dd := t.Date()
		return time.Date(yy, mm, dd, 0, 0, 0, 0, t.Location())
	case Week:
		yy, mm, dd := t.Date()
		day := time.Date(yy, mm, dd, 0, 0, 0, 0, t.Location())
		weekday := int(day.Weekday())
		return day.AddDate(0, 0, -weekday)
	case Month:
		yy, mm, _ := t.Date()
		return time.Date(yy, mm, 1, 0, 0, 0, 0, t.Location())
	}

	return t
}

type Intervals struct {
	keys   []string
	values []*roaring64.Bitmap
	b      roaring64.Bitmap
	i      Interval
	format string
	last   time.Time
}

func NewIntervals(i Interval) *Intervals {
	return &Intervals{i: i, format: i.Format()}
}

func (i *Intervals) Each(f func(date string, r *roaring64.Bitmap)) {
	for n := range i.keys {
		f(i.keys[n], i.values[n])
	}
}

func (i *Intervals) Read(row uint64, value int64) {
	ts := i.i.Reduce(value)
	if !i.last.Equal(ts) {
		if !i.b.IsEmpty() {
			i.keys = append(i.keys, i.last.Format(i.format))
			i.values = append(i.values, i.b.Clone())
		}
		i.b.Clear()
		i.last = ts
	}
	i.b.Add(row)
}
