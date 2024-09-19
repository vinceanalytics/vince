package query

import (
	"time"

	"github.com/vinceanalytics/vince/internal/alicia"
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

func (i Interval) Field() alicia.Field {
	switch i {
	case Minute:
		return alicia.MINUTE
	case Hour:
		return alicia.HOUR
	case Date:
		return alicia.DAY
	case Week:
		return alicia.WEEK
	case Month:
		return alicia.MONTH
	default:
		return alicia.Field(0)
	}
}
