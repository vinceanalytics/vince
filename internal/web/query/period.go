package query

import (
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/util/xtime"
)

func period(str, date string) Period {
	base := dateParse(date)
	last := base.End
	switch str {
	case "realtime":
		end := xtime.Now().Truncate(time.Minute)
		return Period{Start: end.Add(-15 * time.Minute), End: end, Interval: Minute}
	case "day":
		base.Interval = Hour
		return Period{Start: beginOfDay(base.Start), End: last, Interval: Hour}
	case "7d":
		first := last.AddDate(0, 0, -6)
		return Period{Start: first, End: last, Interval: Date}
	case "30d":
		first := last.AddDate(0, 0, -30)
		return Period{Start: first, End: last, Interval: Date}
	case "month":
		last = endOfMonth(last)
		first := beginOfMonth(last)
		return Period{Start: first, End: last, Interval: Date}
	case "6mo":
		last = endOfMonth(last)
		first := beginOfMonth(last.AddDate(0, -5, 0))
		return Period{Start: first, End: last, Interval: Date}
	case "12mo":
		last = endOfMonth(last)
		first := beginOfMonth(last.AddDate(0, -11, 0))
		return Period{Start: first, End: last, Interval: Date}
	case "year":
		last = endOfYear(last)
		first := beginOfYear(last)
		return Period{Start: first, End: last, Interval: Date}
	case "all":
		return Period{Start: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), End: last, Interval: Date}
	default:
		base.Interval = Hour
		return Period{Start: beginOfDay(base.Start), End: last, Interval: Hour}
	}
}

type Period struct {
	Start, End time.Time
	Interval   Interval
}

func dateParse(date string) Period {
	if date == "" {
		now := xtime.Now()
		return Period{Start: now, End: now}
	}
	a, b, ok := strings.Cut(date, ",")
	start, _ := time.Parse(time.DateOnly, a)

	if ok {
		end, _ := time.Parse(time.DateOnly, b)
		return Period{Start: start, End: end}
	}
	return Period{Start: start, End: endOfDay(start)}
}

func beginOfDay(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}

func beginOfMonth(ts time.Time) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
}

func beginOfYear(ts time.Time) time.Time {
	yy, _, _ := ts.Date()
	return time.Date(yy, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func endOfDay(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), ts.Location())
}

func endOfMonth(ts time.Time) time.Time {
	return beginOfMonth(ts).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func endOfYear(ts time.Time) time.Time {
	return beginOfYear(ts).AddDate(1, 0, 0).Add(-time.Nanosecond)
}
