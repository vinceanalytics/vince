package xtime

import "time"

func Now() time.Time {
	return time.Now().UTC()
}

func UnixMilli(v int64) time.Time {
	return time.UnixMilli(int64(v)).UTC()
}
