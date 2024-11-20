package compute

import (
	"iter"
	"slices"
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
)

func Range(re encoding.Resolution, start, end time.Time) iter.Seq2[int64, int64] {
	if re == encoding.Global {
		return func(yield func(int64, int64) bool) {
			yield(0, 0)
		}
	}
	return func(yield func(int64, int64) bool) {
		var from time.Time
		all := slices.Collect(tr(re, start, end))

		// we iterate in revers because tr generates descending sequence
		for _, to := range slices.Backward(all) {
			if from.IsZero() {
				from = to
				continue
			}
			if !yield(from.UnixMilli(), to.UnixMilli()) {
				return
			}
			from = to
		}
	}

}

func tr(e encoding.Resolution, start, end time.Time) iter.Seq[time.Time] {
	switch e {
	case encoding.Minute:
		return ByMinute(start, end)
	case encoding.Hour:
		return ByHour(start, end)
	case encoding.Day:
		return ByDate(start, end)
	case encoding.Week:
		return ByWeek(start, end)
	case encoding.Month:
		return ByMonth(start, end)
	default:
		return ByDate(start, end)
	}
}
func ByMinute(start, end time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		t := end.Truncate(time.Minute).Add(time.Minute)
		for t.After(start) {
			if !yield(t) {
				return
			}
			t = t.Add(-time.Minute)
		}
	}
}

func ByHour(start, end time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		t := end.Truncate(time.Hour).Add(time.Hour)
		for t.After(start) {
			if !yield(t) {
				return
			}
			t = t.Add(-time.Hour)
		}
	}
}

func ByDate(start, end time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		t := Date(end).AddDate(0, 0, 1)
		for t.After(start) {
			if !yield(t) {
				return
			}
			t = t.AddDate(0, 0, -1)
		}
	}
}

func ByWeek(start, end time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		t := Week(end).AddDate(0, 0, 7)
		for t.After(start) {
			if !yield(t) {
				return
			}
			t = t.AddDate(0, 0, -7)
		}
	}
}

func ByMonth(start, end time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		t := Month(end).AddDate(0, 1, 0)
		for t.After(start) {
			if !yield(t) {
				return
			}
			t = t.AddDate(0, -1, 0)
		}
	}
}
