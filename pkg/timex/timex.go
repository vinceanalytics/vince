package timex

import (
	"time"
)

const HumanDate = "Jan 02, 2006"

func beginningOfDay(ts time.Time) time.Time {
	ts = ts.UTC()
	y, m, d := ts.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func beginningOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	return t.AddDate(0, 0, -weekday)
}

func beginningOfMonth(ts time.Time) time.Time {
	y, m, _ := ts.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
}

func beginningOfYear(ts time.Time) time.Time {
	y, _, _ := ts.Date()
	return time.Date(y, time.January, 1, 0, 0, 0, 0, time.UTC)
}
