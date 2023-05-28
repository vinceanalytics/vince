package templates

import (
	"fmt"
	"time"

	"github.com/gernest/vince/pkg/timex"
)

type CalendarEntry struct {
	Name       string
	Start, End time.Time
}

func (c *CalendarEntry) From() string {
	return c.Start.Format(time.RFC3339)
}

func (c *CalendarEntry) To() string {
	return c.End.Format(time.RFC3339)
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
		thisYear(),
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
		Name:  "This month",
		End:   ts,
		Start: timex.BeginningOfMonth(ts),
	}
}

func lastMonth() CalendarEntry {
	ts := timex.BeginningOfMonth(time.Now())
	lastMonthDays := timex.DaysInMonth(ts.Add(-time.Hour * 24))
	return CalendarEntry{
		Name:  "Last month",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * time.Duration(lastMonthDays)),
	}
}

func past7Days() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "Past 7 days",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * 7),
	}
}

func past14Days() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "Past 14 days",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * 14),
	}
}

func past30Days() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "Past 30 days",
		End:   ts,
		Start: ts.Add(-time.Hour * 24 * 30),
	}
}

func thisYear() CalendarEntry {
	ts := timex.Today()
	return CalendarEntry{
		Name:  "This year",
		End:   ts,
		Start: timex.BeginningOfYear(ts),
	}
}

func thisYearFormat() string {
	ts := thisYear()
	return fmt.Sprintf("%s - %s",
		ts.Start.Format(timex.HumanDate),
		ts.End.Format(timex.HumanDate),
	)
}
