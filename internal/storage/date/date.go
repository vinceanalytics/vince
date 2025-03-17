package date

import (
	"fmt"
	"strings"
	"time"
)

const (
	minute   = 60
	hour     = minute * 60
	day      = hour * 24
	nsPerSec = 1e9
	maxUnix  = (1<<16)*day - 1
)

func Resolve(ts time.Time) (mins, hrs, days, weeks, month, year uint64) {
	ts = ts.UTC()
	secs := uint64(ts.Unix())
	mins = secs / minute
	hrs = secs / hour
	days = secs / day

	yy, mm, dd := ts.Date()
	t := time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, 0, -int(t.Weekday()))
	weeks = uint64(t.Unix()) / day
	t = time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
	month = uint64(t.Unix()) / day
	t = time.Date(yy, time.January, 1, 0, 0, 0, 0, time.UTC)
	year = uint64(t.Unix()) / day
	return
}

func Debug(ts time.Time) string {
	mins, hrs, days, weeks, month, year := Resolve(ts)
	var s strings.Builder
	fmt.Fprintln(&s, Minute(mins).Format(time.RubyDate))
	fmt.Fprintln(&s, Hour(hrs).Format(time.RubyDate))
	fmt.Fprintln(&s, Day(days).Format(time.RubyDate))
	fmt.Fprintln(&s, Week(weeks).Format(time.RubyDate))
	fmt.Fprintln(&s, Month(month).Format(time.RubyDate))
	fmt.Fprintln(&s, Year(year).Format(time.RubyDate))
	return s.String()
}

func Minute(v uint64) time.Time {
	return time.Unix(int64(v*minute), 0).UTC()
}

func FromMinute(ts time.Time, add bool) uint64 {
	if add {
		ts = ts.Add(time.Minute)
	}
	return uint64(ts.Unix()) / minute
}

func Hour(v uint64) time.Time {
	return time.Unix(int64(v*hour), 0).UTC()
}

func FromHour(ts time.Time, add bool) uint64 {
	if add {
		ts = ts.Add(time.Hour)
	}
	return uint64(ts.Unix()) / hour
}

func Day(v uint64) time.Time {
	return time.Unix(int64(v*day), 0).UTC()
}

func FromDay(ts time.Time, add bool) uint64 {
	if add {
		ts = ts.AddDate(0, 0, 1)
	}
	return uint64(ts.Unix()) / day
}

func Week(v uint64) time.Time {
	return Day(v)
}

func FromWeek(ts time.Time, add bool) uint64 {
	if add {
		ts = ts.AddDate(0, 0, 7)
	}
	yy, mm, dd := ts.Date()
	t := time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, 0, -int(t.Weekday()))
	return uint64(t.Unix()) / day
}

func Month(v uint64) time.Time {
	return Day(v)
}

func FromMonth(ts time.Time, add bool) uint64 {
	if add {
		ts = ts.AddDate(0, 1, 0)
	}
	yy, mm, _ := ts.Date()
	t := time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
	return uint64(t.Unix()) / day
}

func Year(v uint64) time.Time {
	return Day(v)
}

func FromYear(ts time.Time, add bool) uint64 {
	if add {
		ts = ts.AddDate(1, 0, 0)
	}
	yy, _, _ := ts.Date()
	t := time.Date(yy, time.January, 1, 0, 0, 0, 0, time.UTC)
	return uint64(t.Unix()) / day
}
