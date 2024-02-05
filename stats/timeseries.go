package stats

import (
	"context"
	"slices"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	"github.com/vinceanalytics/staples/staples/filters"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/logger"
	"github.com/vinceanalytics/staples/staples/session"
	"github.com/vinceanalytics/staples/staples/timeutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TimeSeries(ctx context.Context, req *v1.Timeseries_GetTimeseriesRequest) (*v1.Timeseries_GetTimeseriesResponse, error) {
	log := logger.Get(ctx)
	// make sure we have valid interval
	if !ValidByPeriod(req.Period, req.Interval) {
		return nil, status.Error(codes.InvalidArgument, "Interval out of range")
	}
	filters := &v1.Filters{
		List: append(req.Filters, &v1.Filter{
			Property: v1.Property_domain,
			Op:       v1.Filter_equal,
			Value:    req.SiteId,
		}),
	}
	metrics := slices.Clone(req.Metrics)
	slices.Sort(metrics)
	metricsToProjection(filters, metrics)
	from, to := PeriodToRange(time.Now, req.Period)
	r, err := session.Get(ctx).Scan(ctx, from.UnixMilli(), to.UnixMilli(), filters)
	if err != nil {
		log.Error("Failed scanning", "err", err)
		return nil, InternalError
	}

	mapping := map[string]int{}
	for i := 0; i < int(r.NumCols()); i++ {
		mapping[r.ColumnName(i)] = i
	}
	tsKey := mapping[v1.Filters_Timestamp.String()]
	ts := r.Column(tsKey).(*array.Int64).Int64Values()
	var buckets []*v1.Timeseries_Bucket

	err = timeutil.TimeBuckets(req.Interval, ts, func(bucket int64, start, end int) error {
		n := r.NewSlice(int64(start), int64(end))
		defer n.Release()
		buck := &v1.Timeseries_Bucket{
			Timestamp: timestamppb.New(
				time.UnixMilli(bucket),
			),
		}
		var visits *float64
		var view *float64
		for _, x := range metrics {
			var value float64
			switch x {
			case v1.Metric_pageviews:
				a := n.Column(mapping[v1.Filters_Event.String()])
				count := calcPageViews(a)
				view = &count
				value = count
			case v1.Metric_visitors:
				a := n.Column(mapping[v1.Filters_ID.String()])
				u, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(a))
				if err != nil {
					return err
				}
				value = float64(u.Len())
				u.Release()
			case v1.Metric_visits:
				a := n.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
				sum := float64(math.Int64.Sum(a))
				visits = &sum
				value = sum
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
				value = sum
			case v1.Metric_visit_duration:
				a := n.Column(mapping[v1.Filters_Duration.String()]).(*array.Float64)
				sum := float64(math.Float64.Sum(a))
				count := float64(a.Len())
				var avg float64
				if count != 0 {
					avg = sum / count
				}
				value = avg
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
				value = vw
			case v1.Metric_events:
				a := n.Column(mapping[v1.Filters_Event.String()])
				value = float64(a.Len())
			}
			buck.Values = append(buck.Values, &v1.Value{
				Metric: x,
				Value:  value,
			})
			buckets = append(buckets, buck)
		}
		return nil
	})
	if err != nil {
		log.Error("Failed processing buckets", "err", err)
		return nil, InternalError
	}
	return &v1.Timeseries_GetTimeseriesResponse{Results: buckets}, nil
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

func metricsToProjection(f *v1.Filters, me []v1.Metric, props ...v1.Property) {
	m := make(map[v1.Filters_Projection]struct{})
	m[v1.Filters_Timestamp] = struct{}{}
	for _, p := range props {
		m[filters.PropToProjection[p]] = struct{}{}
	}
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
