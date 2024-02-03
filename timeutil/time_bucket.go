package timeutil

import "time"

type Interval uint

const (
	Days Interval = iota
	Week
	Month
	Year
)

func TimeBuckets(interval Interval, source []int64, cb func(bucket int64, start, end int) error) error {
	if len(source) == 0 {
		return nil
	}
	var start int
	bucket := next(source[0], interval)
	for i := 0; i < len(source); i++ {
		if source[i] <= bucket {
			continue
		}
		if err := cb(bucket, start, i); err != nil {
			return err
		}
		start = i
		bucket = next(source[i], interval)
	}
	return cb(bucket, start, len(source))
}

func next(ts int64, i Interval) (v int64) {
	switch i {
	case Days:
		v = EndDay(time.UnixMilli(ts)).UnixMilli()
	case Week:
		v = EndWeek(time.UnixMilli(ts)).UnixMilli()
	case Month:
		v = EndDay(time.UnixMilli(ts)).UnixMilli()
	case Year:
		v = EndYear(time.UnixMilli(ts)).UnixMilli()
	}
	return
}
