package stats

import (
	"time"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/timeutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Avoid leaking internal errors to client. The actual error is logged and this
// is returned back to the client.
var InternalError = status.Error(codes.Internal, "Something went wrong")

func PeriodToRange(now func() time.Time, period *v1.TimePeriod) (start, end time.Time) {
	switch e := period.Value.(type) {
	case *v1.TimePeriod_Base_:
		switch e.Base {
		case v1.TimePeriod_day:
			end = timeutil.Today()
			start = end
		case v1.TimePeriod__7d:
			end = timeutil.Today()
			start = end.AddDate(0, 0, -6)
		case v1.TimePeriod__30d:
			end = timeutil.Today()
			start = end.AddDate(0, 0, -30)
		case v1.TimePeriod_mo:
			end = timeutil.Today()
			start = timeutil.BeginMonth(end)
			end = timeutil.EndMonth(end)
		case v1.TimePeriod__6mo:
			end = timeutil.EndMonth(timeutil.Today())
			start = timeutil.BeginMonth(end.AddDate(0, -5, 0))
		case v1.TimePeriod__12mo:
			end = timeutil.EndMonth(timeutil.Today())
			start = timeutil.BeginMonth(end.AddDate(0, -11, 0))
		case v1.TimePeriod_year:
			end = timeutil.EndYear(timeutil.Today())
			start = timeutil.BeginYear(end)
		}

	case *v1.TimePeriod_Custom_:
		end = e.Custom.End.AsTime()
		start = e.Custom.Start.AsTime()
	}
	return
}

func ValidByPeriod(period *v1.TimePeriod, i v1.Interval) bool {
	switch e := period.Value.(type) {
	case *v1.TimePeriod_Base_:
		switch e.Base {
		case v1.TimePeriod_day:
			return i == v1.Interval_minute || i == v1.Interval_hour
		case v1.TimePeriod__7d:
			return i == v1.Interval_hour || i == v1.Interval_date
		case v1.TimePeriod_mo, v1.TimePeriod__30d:
			return i == v1.Interval_date || i == v1.Interval_week
		case v1.TimePeriod__6mo, v1.TimePeriod__12mo, v1.TimePeriod_year:
			return i == v1.Interval_date || i == v1.Interval_week || i == v1.Interval_month
		default:
			return false
		}
	case *v1.TimePeriod_Custom_:
		return i == v1.Interval_date || i == v1.Interval_week || i == v1.Interval_month
	default:
		return false
	}
}
