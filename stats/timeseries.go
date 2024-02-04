package stats

import (
	"context"
	"slices"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	"github.com/vinceanalytics/staples/staples/db"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/logger"
	"github.com/vinceanalytics/staples/staples/timeutil"
)

func TimeSeries(
	ctx context.Context,
	scanner db.Scanner,
	start, end int64,
	metrics []v1.Metric,
	filters *v1.Filters,
	interval timeutil.Interval,
) arrow.Record {
	log := logger.Get(ctx).With("call", "stats.TimeSeries")
	slices.Sort(metrics)
	metricsToProjection(filters, metrics)
	r, err := scanner.Scan(ctx, start, end, filters)
	if err != nil {
		log.Error("Failed scanning", "err", err)
		return nil
	}
	fields := []arrow.Field{
		{Name: v1.Filters_Timestamp.String(), Type: arrow.PrimitiveTypes.Int64},
	}
	for _, m := range metrics {
		fields = append(fields, arrow.Field{
			Name: m.String(), Type: arrow.PrimitiveTypes.Float64,
		})
	}
	b := array.NewRecordBuilder(compute.GetAllocator(ctx),
		arrow.NewSchema(fields, nil),
	)
	defer b.Release()

	mapping := map[string]int{}
	for i := 0; i < int(r.NumCols()); i++ {
		mapping[r.ColumnName(i)] = i
	}
	ts := r.Column(mapping[v1.Filters_Timestamp.String()]).(*array.Int64).Int64Values()
	err = timeutil.TimeBuckets(interval, ts, func(bucket int64, start, end int) error {
		n := r.NewSlice(int64(start), int64(end))
		defer n.Release()
		b.Field(0).(*array.Int64Builder).Append(bucket)
		for i, x := range metrics {
			var visits *float64
			var view *float64
			switch x {
			case v1.Metric_pageviews:
				a := n.Column(mapping[v1.Filters_Event.String()])
				count := calcPageViews(a)
				view = &count
				b.Field(i + 1).(*array.Float64Builder).Append(count)
			case v1.Metric_visitors:
				a := n.Column(mapping[v1.Filters_ID.String()])
				u, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(a))
				if err != nil {
					return err
				}
				b.Field(i + 1).(*array.Float64Builder).Append(float64(u.Len()))
				u.Release()
			case v1.Metric_visits:
				a := n.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
				sum := float64(math.Int64.Sum(a))
				visits = &sum
				b.Field(i + 1).(*array.Float64Builder).Append(sum)
			case v1.Metric_bounce_rate:
				var vis float64
				if visits != nil {
					vis = *visits
				} else {
					a := n.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
					vis = float64(math.Int64.Sum(a))
				}
				a := n.Column(mapping[v1.Filters_Bounce.String()]).(*array.Int64)
				sum := float64(math.Int64.Sum(a))
				if vis != 0 {
					sum /= vis
				}
				b.Field(i + 1).(*array.Float64Builder).Append(sum)
			case v1.Metric_visit_duration:
				a := n.Column(mapping[v1.Filters_Duration.String()]).(*array.Float64)
				sum := float64(math.Float64.Sum(a))
				b.Field(i + 1).(*array.Float64Builder).Append(sum)
			case v1.Metric_views_per_visit:
				var vis float64
				if visits != nil {
					vis = *visits
				} else {
					a := n.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
					vis = float64(math.Int64.Sum(a))
				}
				var vw float64
				if view != nil {
					vw = *view
				} else {
					a := n.Column(mapping[v1.Filters_Event.String()])
					vw = calcPageViews(a)
				}
				if vis != 0 {
					vw /= vis
				}
				b.Field(i + 1).(*array.Float64Builder).Append(vw)
			case v1.Metric_events:
				a := n.Column(mapping[v1.Filters_Event.String()])
				b.Field(i + 1).(*array.Float64Builder).Append(float64(a.Len()))
			}
		}
		return nil
	})
	if err != nil {
		log.Error("Failed processing buckets", "err", err)
		return nil
	}
	return b.NewRecord()
}

const viewStr = "pageview"

func calcPageViews(a arrow.Array) (n float64) {
	d := a.(*array.Dictionary)
	x := d.Dictionary().(*array.String)
	for i := 0; i < d.Len(); i++ {
		if d.IsNull(i) {
			continue
		}
		if x.Value(d.GetValueIndex(i)) == viewStr {
			n++
		}
	}
	return
}

func buckets(r arrow.Record, f func(int64, arrow.Record) error) error {
	return nil
}

func metricsToProjection(f *v1.Filters, me []v1.Metric) {
	m := make(map[v1.Filters_Projection]struct{})
	m[v1.Filters_Timestamp] = struct{}{}
	for _, v := range me {
		switch v {
		case v1.Metric_pageviews:
			m[v1.Filters_Event] = struct{}{}
		case v1.Metric_visitors:
			m[v1.Filters_ID] = struct{}{}
		case v1.Metric_visits:
			m[v1.Filters_Session] = struct{}{}
		case v1.Metric_bounce_rate:
			m[v1.Filters_Session] = struct{}{}
			m[v1.Filters_Bounce] = struct{}{}
		case v1.Metric_visit_duration:
			m[v1.Filters_Duration] = struct{}{}
		case v1.Metric_views_per_visit:
			m[v1.Filters_Event] = struct{}{}
			m[v1.Filters_Duration] = struct{}{}
		}
	}
	for k := range m {
		f.Projection = append(f.Projection, k)
	}
}
