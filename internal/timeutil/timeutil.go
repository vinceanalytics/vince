package timeutil

import "time"

func Today() time.Time {
	return BeginDay(time.Now())
}

func beginMinute(ts time.Time) time.Time {
	return ts.Truncate(time.Minute)
}

func beginHour(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, ts.Hour(), 0, 0, 0, ts.Location())
}

func BeginDay(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, ts.Location())
}

func Day(ts time.Time, dd int) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, ts.Location())
}

func BeginWeek(ts time.Time) time.Time {
	ts = BeginDay(ts)
	return ts.AddDate(0, 0, -int(ts.Weekday()))
}

func BeginMonth(ts time.Time) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, ts.Location())
}

func EndOfHour(ts time.Time) time.Time {
	return beginHour(ts).Add(time.Hour - time.Nanosecond)
}

func EndOfMinute(ts time.Time) time.Time {
	return beginMinute(ts).Add(time.Minute - time.Nanosecond)
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
	return time.Date(yy, time.January, 1, 0, 0, 0, 0, ts.Location())
}

func EndYear(ts time.Time) time.Time {
	return BeginYear(ts).AddDate(1, 0, 0).Add(-time.Nanosecond)
}
