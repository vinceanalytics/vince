package timeutil

import (
	"time"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
)

func TimeBuckets(interval v1.Interval, source []int64, cb func(bucket int64, start, end int) error) error {
	if len(source) == 0 {
		return nil
	}
	var start int
	bucket := hash(source[0], interval)
	for i := 0; i < len(source); i++ {
		h := hash(source[i], interval)
		if h == bucket {
			continue
		}
		if err := cb(bucket, start, i); err != nil {
			return err
		}
		start = i
		bucket = h
	}
	return cb(bucket, start, len(source))
}

func hash(ts int64, i v1.Interval) (v int64) {
	switch i {
	case v1.Interval_minute:
		v = EndOfMinute(time.UnixMilli(ts)).UnixMilli()
	case v1.Interval_hour:
		v = EndOfHour(time.UnixMilli(ts)).UnixMilli()
	case v1.Interval_date:
		v = EndDay(time.UnixMilli(ts)).UnixMilli()
	case v1.Interval_week:
		v = EndWeek(time.UnixMilli(ts)).UnixMilli()
	case v1.Interval_month:
		v = EndDay(time.UnixMilli(ts)).UnixMilli()
	}
	return
}
