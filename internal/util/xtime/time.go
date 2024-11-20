package xtime

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func UnixMilli(v int64) time.Time {
	return time.UnixMilli(int64(v)).UTC()
}

func Test() time.Time {
	ts, _ := time.Parse(time.RFC822, time.RFC822)
	return ts.UTC()
}
