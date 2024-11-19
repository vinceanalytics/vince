package compute

import (
	"iter"
	"time"
)

func ByMinute(start, end time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		t := end.Truncate(time.Minute)
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
		t := end.Truncate(time.Hour)
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
		t := Date(end)
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
		t := Week(end)
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
		t := Month(end)
		for t.After(start) {
			if !yield(t) {
				return
			}
			t = t.AddDate(0, -1, 0)
		}
	}
}
