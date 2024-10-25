package query

import (
	"iter"
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

func (i Interval) Range(start, end time.Time) iter.Seq[time.Time] {
	switch i {
	case Minute:
		return compute.ByMinute(start, end)
	case Hour:
		return compute.ByHour(start, end)
	case Date:
		return compute.ByDate(start, end)
	case Week:
		return compute.ByWeek(start, end)
	case Month:
		return compute.ByMonth(start, end)
	default:
		return compute.ByDate(start, end)
	}
}
