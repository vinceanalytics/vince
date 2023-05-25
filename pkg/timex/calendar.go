package timex

import (
	"time"
)

type Range struct {
	From, To time.Time
}

func (r *Range) TS() time.Time {
	if r.From.IsZero() {
		return r.To
	}
	if r.To.IsZero() {
		return r.From
	}
	return r.From
}

func (r Range) Build() (o []Range) {
	diff := r.To.Year() - r.From.Year()
	if diff == 0 {
		return []Range{r}
	}
	ts := BeginningOfYear(r.From)
	for i := 0; i < diff; i += 1 {
		if i == 0 {
			// cover full year beginning from
			// from ...
			o = append(o, Range{
				From: r.From,
			})
			continue
		}
		if i == diff-1 {
			// last year cover full year up to end
			// ... to
			o = append(o, Range{
				To: r.To,
			})
			continue
		}
		ts = ts.AddDate(i, 0, 0)
		// full calendar year
		o = append(o, Range{
			From: BeginningOfYear(ts),
			To:   EndOfYear(ts),
		})
	}
	return
}

func (r *Range) Timestamps() (o []int64) {
	for _, b := range r.Build() {
		o = append(o, b.timestamps()...)
	}
	return
}

func (r *Range) timestamps() []int64 {
	ts := r.TS()
	start := r.From
	if start.IsZero() {
		start = BeginningOfYear(ts)
	}
	end := r.To
	if end.IsZero() {
		end = EndOfYear(ts)
	}
	start = start.Truncate(time.Hour)
	end = end.Truncate(time.Hour)
	diff := end.Sub(start).Truncate(time.Hour).Hours()
	ls := make([]int64, int(diff+1))
	for i := range ls {
		ls[i] = start.Add(time.Duration(i) * time.Hour).Unix()
	}
	return ls
}

// CalendarHours returns total hours in a calendar year.
func CalendarHours(now time.Time) int {
	ts := BeginningOfYear(now).Truncate(time.Hour)
	end := EndOfYear(now).Truncate(time.Hour)
	diff := end.Sub(ts).Truncate(time.Hour).Hours()
	return int(diff + 1)
}
