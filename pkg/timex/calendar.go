package timex

import (
	"time"
)

// CalendarHours returns total hours in a calendar year represented by ts
func CalendarHours(ts time.Time) int {
	return DaysInAYear(ts) * 24
}

// HourIndex returns the index in calendar year array where the hour for ts is
// found.
func HourIndex(ts time.Time) int {
	day := ts.YearDay()
	if day == 0 {
		return ts.Hour()
	}
	// count all hours before today and add today's hours.
	return ((day - 1) * 24) + ts.Hour()
}
