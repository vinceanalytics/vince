package db

import (
	"context"
	"slices"
	"time"

	"github.com/gernest/rbf/dsl/bsi"
	"github.com/gernest/rbf/dsl/tx"
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
	a := &timeseriesQuery{metrics: m}
	from, to := periodToRange(req.Period, req.Date)
	props := append(req.Filters,
		&v1.Filter{Property: v1.Property_domain, Op: v1.Filter_equal, Value: req.SiteId},
	)
	ts := bsi.Filter("timestamp", bsi.RANGE, from.UnixMilli(), to.UnixMilli())
	fs := filterProperties(props...)

	r, err := db.db.Reader()
	if err != nil {
		return nil, err
	}
	defer r.Release()
	for _, shard := range r.Range(from, to) {
		err := r.View(shard, func(txn *tx.Tx) error {
			f, err := ts.Apply(txn, nil)
			if err != nil {
				return err
			}
			if f.IsEmpty() {
				return nil
			}
			r, err := fs.Apply(txn, f)
			if err != nil {
				return err
			}
			if r.IsEmpty() {
				return nil
			}
			return a.View(txn.View).Apply(txn, r)
		})
		if err != nil {
			return nil, err
		}
	}
	return &v1.Timeseries_Response{Results: a.Result(req.Interval)}, nil
}

type timeseriesQuery struct {
	series  map[time.Time]*aggregate
	metrics []v1.Metric
}

func (t *timeseriesQuery) View(view string) View {
	ts, _ := time.Parse(layout, view)
	a, ok := t.series[ts]
	if !ok {
		a = newAggregate(t.metrics)
		t.series[ts] = a
	}
	return a
}

func (t *timeseriesQuery) Result(interval v1.Interval) []*v1.Timeseries_Bucket {
	switch interval {
	default:
		keys := make([]time.Time, 0, len(t.series))
		o := make([]*v1.Timeseries_Bucket, 0, len(t.series))
		for k := range t.series {
			keys = append(keys, k)
		}
		slices.SortFunc(keys, func(a, b time.Time) int {
			return a.Compare(b)
		})
		for i := range keys {
			x := &v1.Timeseries_Bucket{
				Timestamp: timestamppb.New(keys[i]),
				Values:    make(map[string]float64),
			}
			a := t.series[keys[i]]
			for _, m := range t.metrics {
				x.Values[m.String()] = a.Result(m)
			}
			o = append(o, x)
		}
		return o
	}
}
