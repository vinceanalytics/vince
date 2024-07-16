package db

import (
	"context"
	"slices"
	"time"

	"github.com/gernest/rows"
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
	a := &timeseriesQuery{metrics: m, series: make(map[uint64]*aggregate)}
	from, to := periodToRange(req.Period, req.Date)
	err = db.view(func(tx *view) error {
		shards := db.shards.Iterator()
		for shards.HasNext() {
			shard := shards.Next()
			r, err := tx.domain(shard, req.SiteId)
			if err != nil {
				return err
			}
			if r.IsEmpty() {
				continue
			}
			r, err = tx.time(shard, from, to, r)
			if err != nil {
				return err
			}
			if r.IsEmpty() {
				continue
			}
			err = a.Apply(tx, shard, r)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &v1.Timeseries_Response{Results: a.Result()}, nil
}

type timeseriesQuery struct {
	series  map[uint64]*aggregate
	metrics []v1.Metric
}

func (t *timeseriesQuery) Apply(tx *view, shard uint64, f *rows.Row) error {
	uniq, err := tx.unique("date", shard, f)
	if err != nil {
		return err
	}
	for ts, rws := range uniq {

		g, ok := t.series[ts]
		if !ok {
			g = newAggregate(t.metrics)
			t.series[ts] = g
		}
		err = g.Apply(tx, shard, rows.NewRow(rws...))
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *timeseriesQuery) Result() []*v1.Timeseries_Bucket {
	keys := make([]uint64, 0, len(t.series))
	o := make([]*v1.Timeseries_Bucket, 0, len(t.series))
	for k := range t.series {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for i := range keys {
		x := &v1.Timeseries_Bucket{
			Timestamp: timestamppb.New(time.UnixMilli(int64(keys[i]))),
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
