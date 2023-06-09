package timex

import (
	"time"
)

type Duration uint8

const (
	Unknown Duration = iota
	Today
	ThisWeek
	ThisMonth
	ThisYear
)

var (
	_cal_name = map[Duration]string{
		Today:     "Today",
		ThisWeek:  "This week",
		ThisMonth: "This month",
		ThisYear:  "This year",
	}
	_cal_value = map[string]Duration{
		"Today":      Today,
		"This week":  ThisWeek,
		"This month": ThisMonth,
		"This year":  ThisYear,
	}
)

func Parse(s string) Duration {
	v, ok := _cal_value[s]
	if !ok {
		return Today
	}
	return v
}

func (c Duration) String() string {
	return _cal_name[c]
}

func (c Duration) Window(ts time.Time) time.Duration {
	switch c {
	case Unknown, Today:
		return time.Hour * 24
	case ThisWeek:
		return ts.Sub(beginningOfWeek(ts))
	case ThisMonth:
		return ts.Sub(beginningOfMonth(ts))
	case ThisYear:
		return ts.Sub(beginningOfYear(ts))
	default:
		return time.Duration(0)
	}
}
