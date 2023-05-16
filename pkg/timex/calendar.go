package timex

import (
	"time"
)

// CalendarHours returns total hours in a calendar year.
func CalendarHours(now time.Time) int {
	ts := BeginningOfYear(now).Truncate(time.Hour)
	end := EndOfYear(now).Truncate(time.Hour)
	diff := end.Sub(ts).Truncate(time.Hour).Hours()
	return int(diff + 1)
}

func YearTimestamps(now time.Time) []int64 {
	ts := BeginningOfYear(now).Truncate(time.Hour)
	ls := make([]int64, CalendarHours(now))
	for i := range ls {
		ls[i] = ts.Add(time.Duration(i) * time.Hour).Unix()
	}
	return ls
}

// HourIndex returns the index in calendar year array where the hour for ts is
// found.
func HourIndex(ts time.Time) int {
	begin := BeginningOfYear(ts).Truncate(time.Hour)
	ts = ts.Truncate(time.Hour)
	return int(ts.Sub(begin).Truncate(time.Hour).Hours())
}
