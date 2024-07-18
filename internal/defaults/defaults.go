package defaults

import (
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/timeutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Set populates ms with default values if they are not yet set. This ensures
// requests are in correct state before validation
func Set(msg proto.Message) {
	switch e := msg.(type) {
	case *v1.Realtime_Request:

	case *v1.Aggregate_Request:

		if e.Period == nil {
			e.Period = &v1.TimePeriod{Value: &v1.TimePeriod_Base_{Base: v1.TimePeriod__30d}}
		}
		if len(e.Metrics) == 0 {
			e.Metrics = []v1.Metric{v1.Metric_visitors}
		}
		if e.Date == nil {
			e.Date = timestamppb.New(timeutil.EndDay(time.Now()))
		}
	case *v1.Timeseries_Request:

		if e.Period == nil {
			e.Period = &v1.TimePeriod{Value: &v1.TimePeriod_Base_{Base: v1.TimePeriod__30d}}
		}
		if len(e.Metrics) == 0 {
			e.Metrics = []v1.Metric{v1.Metric_visitors}
		}
		if e.Date == nil {
			e.Date = timestamppb.New(timeutil.EndDay(time.Now()))
		}
	case *v1.BreakDown_Request:
		if e.Period == nil {
			e.Period = &v1.TimePeriod{Value: &v1.TimePeriod_Base_{Base: v1.TimePeriod__30d}}
		}
		if len(e.Metrics) == 0 {
			e.Metrics = []v1.Metric{v1.Metric_visitors}
		}
		if e.Date == nil {
			e.Date = timestamppb.New(timeutil.EndDay(time.Now()))
		}
	}
}
