package query

import (
	"time"

	"github.com/vinceanalytics/vince/internal/compute"
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

func (i Interval) Range(start, end time.Time, f func(time.Time) error) error {
	switch i {
	case Minute:
		return compute.ByMinute(start, end, f)
	case Hour:
		return compute.ByHour(start, end, f)
	case Date:
		return compute.ByDate(start, end, f)
	case Week:
		return compute.ByWeek(start, end, f)
	case Month:
		return compute.ByMonth(start, end, f)
	default:
		return compute.ByDate(start, end, f)
	}
}
