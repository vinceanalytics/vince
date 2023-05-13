// Package timex expose time utility functions. All operations on this package
// first converts time.Time to UTC.
//
// Handling timezones is a complex affair. By coercing to UTC instead of assuming
// time is UTC allows the user to decide on what time they want, we then convert to what
// we want before using it.
package timex

import (
	"time"

	"github.com/jinzhu/now"
)

const HumanDate = "Jan 02, 2006"

func EndOfDay(ts time.Time) time.Time {
	ts = ts.UTC()
	y, m, d := ts.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
}

func BeginningOfDay(ts time.Time) time.Time {
	ts = ts.UTC()
	y, m, d := ts.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func Date(ts time.Time) time.Time {
	ts = ts.UTC()
	y, m, d := ts.Date()
	return time.Date(
		y, m, d, 0, 0, 0, 0, ts.Location(),
	)
}

func Today() time.Time {
	return Date(time.Now())
}

func BeginningOfMonth(ts time.Time) time.Time {
	ts = ts.UTC()
	y, m, _ := ts.Date()
	now.BeginningOfHour()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

func BeginningOfHour(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, ts.Hour(), 0, 0, 0, time.UTC)
}

func BeginningOfYear(ts time.Time) time.Time {
	ts = ts.UTC()
	y, _, _ := ts.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func EndOfMonth(ts time.Time) time.Time {
	return BeginningOfMonth(ts).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func EndOfYear(ts time.Time) time.Time {
	return BeginningOfYear(ts).AddDate(1, 0, 0).Add(-time.Nanosecond)
}

func DaysInMonth(ts time.Time) int {
	return EndOfMonth(ts).Day()
}

func DaysInAYear(ts time.Time) int {
	return EndOfYear(ts).YearDay()
}
