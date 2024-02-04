package stats

import (
	"time"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/timeutil"
)

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
