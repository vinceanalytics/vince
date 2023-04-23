package templates

import (
	"time"

	"github.com/gernest/vince/timex"
)

type CalendarEntry struct {
	Name       string
	Start, End time.Time
}

func CalendarEntries() []CalendarEntry {
	return []CalendarEntry{
		today(),
		yesterday(),
		monthToDate(),
		lastMonth(),
		past7Days(),
		past14Days(),
		past30Days(),
	}
}

func today() CalendarEntry {
	ts := timex.BeginningOfDay(time.Now())
	return CalendarEntry{
		Name:  "today",
		Start: ts,
		End:   ts.Add(time.Hour * 24),
	}
}

func yesterday() CalendarEntry {
	ts := timex.BeginningOfDay(time.Now())
	return CalendarEntry{
		Name:  "yesterday",
		End:   ts,
		Start: ts.Add(-time.Hour * 24),
	}
}

func monthToDate() CalendarEntry {
	ts := time.Now()
	return CalendarEntry{
		Name:  "monthToDate",
		End:   ts,
		Start: timex.BeginningOfMonth(ts),
	}
}

func lastMonth() CalendarEntry {
	ts := timex.BeginningOfMonth(time.Now())
	lastMonthDays := timex.DaysInMonth(ts.Add(-time.Hour * 24))
	return CalendarEntry{
		Name:  "lastMonth",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * time.Duration(lastMonthDays)),
	}
}

func past7Days() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "past7Days",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * 7),
	}
}

func past14Days() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "past14Days",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * 14),
	}
}

func past30Days() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "past14Days",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * 30),
	}
}
