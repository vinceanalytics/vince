package timex

import (
	"testing"
	"time"
)

func TestHourIndex(t *testing.T) {
	ts := time.Now()
	n := BeginningOfYear(ts).AddDate(0, 0, 1)
	index := HourIndex(n)
	if index != 24 {
		t.Errorf("expected 24 got %d", index)
	}
	calendarHrs := CalendarHours(ts)
	lastHrIndex := HourIndex(EndOfYear(ts))
	if calendarHrs != lastHrIndex+1 {
		t.Errorf("expected %d == %d", calendarHrs, lastHrIndex+1)
	}
}
