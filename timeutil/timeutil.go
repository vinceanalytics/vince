package timeutil

import "time"

func Today() time.Time {
	return BeginDay(time.Now().UTC())
}

func BeginDay(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}

func Day(ts time.Time, dd int) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}

func BeginWeek(ts time.Time) time.Time {
	ts = BeginDay(ts)
	return ts.AddDate(0, 0, -int(ts.Weekday()))
}

func BeginMonth(ts time.Time) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
}

func EndOfHour(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	hh := ts.Hour()
	return time.Date(yy, mm, dd, hh, 59, 0, 0, time.UTC)
}

func EndOfMinute(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	hh := ts.Hour()
	xx := ts.Minute()
	return time.Date(yy, mm, dd, hh, xx, 59, 0, time.UTC)
}

func EndDay(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 23, 59, 59, int(time.Second-time.Nanosecond), ts.Location())
}

func EndWeek(ts time.Time) time.Time {
	return BeginWeek(ts).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

func EndMonth(ts time.Time) time.Time {
	return BeginMonth(ts).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

func BeginYear(ts time.Time) time.Time {
	yy, _, _ := ts.Date()
	return time.Date(yy, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func EndYear(ts time.Time) time.Time {
	return BeginYear(ts).AddDate(1, 0, 0).Add(-time.Nanosecond)
}
