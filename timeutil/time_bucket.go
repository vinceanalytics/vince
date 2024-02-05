package timeutil

import (
	"time"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/logger"
)

func TimeBuckets(interval v1.Interval, source []int64, cb func(bucket int64, start, end int) error) error {
	if len(source) == 0 {
		return nil
	}
	var start int
	hash := call(interval)
	bucket := hash(source[0])
	for i := 0; i < len(source); i++ {
		h := hash(source[i])
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

func call(i v1.Interval) func(ts int64) int64 {
	switch i {
	case v1.Interval_minute:
		return func(ts int64) int64 {
			return EndOfMinute(time.UnixMilli(ts)).UnixMilli()
		}
	case v1.Interval_hour:
		return func(ts int64) int64 {
			return EndOfHour(time.UnixMilli(ts)).UnixMilli()
		}
	case v1.Interval_date:
		return func(ts int64) int64 {
			return EndDay(time.UnixMilli(ts)).UnixMilli()
		}
	case v1.Interval_week:
		return func(ts int64) int64 {
			return EndWeek(time.UnixMilli(ts)).UnixMilli()
		}
	case v1.Interval_month:
		return func(ts int64) int64 {
			return EndDay(time.UnixMilli(ts)).UnixMilli()
		}
	default:
		logger.Fail("Unexpected interval", "interval", i.String())
		return nil
	}
}
