package db

import (
	"context"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/defaults"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (db *DB) Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error) {
	defaults.Set(req)
	err := validate.Validate(req)
	if err != nil {
		return nil, err
	}
	m := dupe(req.Metrics)
	query := &timeseriesQuery{metrics: m}
	from, to := periodToRange(req.Period, req.Date)
	err = db.Search(from, to, append(req.Filters, &v1.Filter{
		Property: v1.Property_domain,
		Op:       v1.Filter_equal,
		Value:    req.SiteId,
	}), query)
	if err != nil {
		return nil, err
	}
	return &v1.Timeseries_Response{Results: query.Result(req.Interval)}, nil
}

type timeseriesQuery struct {
	series  []*aggregate
	ts      []time.Time
	metrics []v1.Metric
}

var _ Query = (*timeseriesQuery)(nil)

func (t *timeseriesQuery) View(ts time.Time) View {
	a := newAggregate(t.metrics)
	t.series = append(t.series, a)
	t.ts = append(t.ts, ts)
	return a
}

func (t *timeseriesQuery) Result(interval v1.Interval) []*v1.Timeseries_Bucket {
	switch interval {
	default:
		o := make([]*v1.Timeseries_Bucket, 0, len(t.ts))
		for i := range t.ts {
			x := &v1.Timeseries_Bucket{
				Timestamp: timestamppb.New(t.ts[i]),
				Values:    make(map[string]float64),
			}
			a := t.series[i]
			for _, m := range t.metrics {
				x.Values[m.String()] = a.Result(m)
			}
			o = append(o, x)
		}
		return o
	}
}
