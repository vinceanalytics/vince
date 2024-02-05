package stats

import (
	"context"
	"slices"
	"time"

	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/logger"
	"github.com/vinceanalytics/staples/staples/session"
)

func Aggregate(ctx context.Context, req *v1.Aggregate_GetAggegateRequest) (*v1.Aggregate_GetAggregateResponse, error) {
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
		logger.Get(ctx).Error("Failed scanning", "err", err)
		return nil, InternalError
	}
	defer r.Release()
	mapping := map[string]int{}
	for i := 0; i < int(r.NumCols()); i++ {
		mapping[r.ColumnName(i)] = i
	}
	var result []*v1.Value
	var visits *float64
	var view *float64
	for _, metric := range metrics {
		var value float64
		switch metric {
		case v1.Metric_pageviews:
			a := r.Column(mapping[v1.Filters_Event.String()])
			count := calcPageViews(a)
			view = &count
			value = count
		case v1.Metric_visitors:
			a := r.Column(mapping[v1.Filters_ID.String()])
			u, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(a))
			if err != nil {
				logger.Get(ctx).Error("Failed calculating visitors", "err", err)
				return nil, InternalError
			}
			value = float64(u.Len())
			u.Release()
		case v1.Metric_visits:
			a := r.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
			sum := float64(math.Int64.Sum(a))
			visits = &sum
			value = sum
		case v1.Metric_bounce_rate:
			var vis float64
			if visits != nil {
				vis = *visits
			} else {
				a := r.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
				vis = float64(math.Int64.Sum(a))
			}
			a := r.Column(mapping[v1.Filters_Bounce.String()]).(*array.Int64)
			sum := float64(math.Int64.Sum(a))
			if vis != 0 {
				sum /= vis
			}
			value = sum
		case v1.Metric_visit_duration:
			a := r.Column(mapping[v1.Filters_Duration.String()]).(*array.Float64)
			sum := math.Float64.Sum(a)
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
				a := r.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
				vis = float64(math.Int64.Sum(a))
			}
			var vw float64
			if view != nil {
				vw = *view
			} else {
				a := r.Column(mapping[v1.Filters_Event.String()])
				vw = calcPageViews(a)
			}
			if vis != 0 {
				vw /= vis
			}
			value = vw
		case v1.Metric_events:
			a := r.Column(mapping[v1.Filters_Event.String()])
			value = float64(a.Len())
		}
		result = append(result, &v1.Value{
			Metric: metric,
			Value:  value,
		})
	}
	return &v1.Aggregate_GetAggregateResponse{Results: result}, nil
}
