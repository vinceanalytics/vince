package stats

import (
	"net/http"
	"slices"
	"time"

	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/arrow/math"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/request"
	"github.com/vinceanalytics/vince/session"
)

func Aggregate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	req := v1.Aggregate_Request{
		SiteId:  query.Get("site_id"),
		Period:  ParsePeriod(ctx, query),
		Metrics: ParseMetrics(ctx, query),
	}
	if !request.Validate(ctx, w, &req) {
		return
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
	from, to := PeriodToRange(ctx, time.Now, req.Period, r.URL.Query())
	resultRecord, err := session.Get(ctx).Scan(ctx, from.UnixMilli(), to.UnixMilli(), filters)
	if err != nil {
		logger.Get(ctx).Error("Failed scanning", "err", err)
		request.Internal(ctx, w)
		return
	}
	defer resultRecord.Release()
	mapping := map[string]int{}
	for i := 0; i < int(resultRecord.NumCols()); i++ {
		mapping[resultRecord.ColumnName(i)] = i
	}
	var result []*v1.Value
	var visits *float64
	var view *float64
	for _, metric := range metrics {
		var value float64
		switch metric {
		case v1.Metric_pageviews:
			a := resultRecord.Column(mapping[v1.Filters_Event.String()])
			count := calcPageViews(a)
			view = &count
			value = count
		case v1.Metric_visitors:
			a := resultRecord.Column(mapping[v1.Filters_ID.String()])
			u, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(a))
			if err != nil {
				logger.Get(ctx).Error("Failed calculating visitors", "err", err)
				request.Internal(ctx, w)
				return
			}
			value = float64(u.Len())
			u.Release()
		case v1.Metric_visits:
			a := resultRecord.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
			sum := float64(math.Int64.Sum(a))
			visits = &sum
			value = sum
		case v1.Metric_bounce_rate:
			var vis float64
			if visits != nil {
				vis = *visits
			} else {
				a := resultRecord.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
				vis = float64(math.Int64.Sum(a))
			}
			a := resultRecord.Column(mapping[v1.Filters_Bounce.String()]).(*array.Int64)
			sum := float64(math.Int64.Sum(a))
			if vis != 0 {
				sum /= vis
			}
			value = sum
		case v1.Metric_visit_duration:
			a := resultRecord.Column(mapping[v1.Filters_Duration.String()]).(*array.Float64)
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
				a := resultRecord.Column(mapping[v1.Filters_Session.String()]).(*array.Int64)
				vis = float64(math.Int64.Sum(a))
			}
			var vw float64
			if view != nil {
				vw = *view
			} else {
				a := resultRecord.Column(mapping[v1.Filters_Event.String()])
				vw = calcPageViews(a)
			}
			if vis != 0 {
				vw /= vis
			}
			value = vw
		case v1.Metric_events:
			a := resultRecord.Column(mapping[v1.Filters_Event.String()])
			value = float64(a.Len())
		}
		result = append(result, &v1.Value{
			Metric: metric,
			Value:  value,
		})
	}
	request.Write(ctx, w, &v1.Aggregate_Response{Results: result})
	return
}
