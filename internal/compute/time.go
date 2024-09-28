package compute

import "time"

func Hour(ts time.Time) time.Time {
	return ts.Truncate(time.Hour)
}

func Date(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
}

func Week(ts time.Time) time.Time {
	yy, mm, dd := ts.Date()
	t := time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC)
	t = t.AddDate(0, 0, -int(t.Weekday()))
	return t
}

func Month(ts time.Time) time.Time {
	yy, mm, _ := ts.Date()
	return time.Date(yy, mm, 1, 0, 0, 0, 0, time.UTC)
}
