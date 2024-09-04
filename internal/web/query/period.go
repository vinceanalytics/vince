package query

import (
	"strings"
	"time"
)

func period(str, date string) Date {
	base := dateParse(date)
	last := base.Start
	switch str {
	case "realtime":
		end := time.Now().UTC()
		return Date{Start: end.Add(-15 * time.Minute), End: end, Interval: Minute}
	case "day":
		base.Interval = Hour
		return base
	case "7d":
		first := last.AddDate(0, 0, -6)
		return Date{Start: first, End: last, Interval: Day}
	case "30d":
		first := last.AddDate(0, 0, -30)
		return Date{Start: first, End: last, Interval: Day}
	case "month":
		last = endOfMonth(last)
		first := beginOfMonth(last)
		return Date{Start: first, End: last, Interval: Day}
	case "6mo":
		last = endOfMonth(last)
		first := beginOfMonth(last.AddDate(0, -5, 0))
		return Date{Start: first, End: last, Interval: Month}
	case "12mo":
		last = endOfMonth(last)
		first := beginOfMonth(last.AddDate(0, -11, 0))
		return Date{Start: first, End: last, Interval: Month}
	case "year":
		last = endOfYear(last)
		first := beginOfYear(last)
		return Date{Start: first, End: last, Interval: Month}
	case "all":
		return Date{End: last, Interval: Day}
	default:
		base.Interval = Day
		return base
	}
}

type Date struct {
	Start, End time.Time
	Interval   Interval
}

func dateParse(date string) Date {
	if date == "" {
		now := time.Now().UTC()
		return Date{Start: now, End: now}
	}
	a, b, ok := strings.Cut(date, ",")
	start, _ := time.Parse(time.DateOnly, a)

	if ok {
		end, _ := time.Parse(time.DateOnly, b)
		return Date{Start: start, End: end}
	}
	return Date{Start: start, End: start}
}

func beginOfMonth(ts time.Time) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
}

func beginOfYear(ts time.Time) time.Time {
	yy, _, _ := ts.Date()
	return time.Date(yy, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func endOfMonth(ts time.Time) time.Time {
	return beginOfMonth(ts).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func endOfYear(ts time.Time) time.Time {
	return beginOfMonth(ts).AddDate(1, 0, 0).Add(-time.Nanosecond)
}
