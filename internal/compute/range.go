package compute

import "time"

func ByMinute(start, end time.Time, f func(time.Time) error) error {
	t := start.Truncate(time.Minute)
	for t.Before(end) {
		err := f(t)
		if err != nil {
			return err
		}
		t = t.Add(time.Minute)
	}
	return nil
}

func ByHour(start, end time.Time, f func(time.Time) error) error {
	t := start.Truncate(time.Hour)
	for t.Before(end) {
		err := f(t)
		if err != nil {
			return err
		}
		t = t.Add(time.Hour)
	}
	return nil
}

func ByDate(start, end time.Time, f func(time.Time) error) error {
	t := Date(start)
	for t.Before(end) {
		err := f(t)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 0, 1)
	}
	return nil
}

func ByWeek(start, end time.Time, f func(time.Time) error) error {
	t := Week(start)
	for t.Before(end) {
		err := f(t)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 0, 7)
	}
	return nil
}

func ByMonth(start, end time.Time, f func(time.Time) error) error {
	t := Month(start)
	for t.Before(end) {
		err := f(t)
		if err != nil {
			return err
		}
		t = t.AddDate(0, 1, 0)
	}
	return nil
}
