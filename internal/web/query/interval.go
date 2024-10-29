package query

import (
	"time"
)

type Interval byte

const (
	Minute Interval = 1 + iota
	Hour
	Date
	Week
	Month
)

func (i Interval) Format() string {
	switch i {
	case Minute, Hour, Date:
		return time.DateTime
	default:
		return time.DateOnly
	}
}

func (i Interval) String() string {
	switch i {
	case Minute:
		return "minute"
	case Hour:
		return "hour"
	case Date:
		return "date"
	case Week:
		return "week"
	case Month:
		return "month"
	default:
		return ""
	}
}
