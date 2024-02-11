package stats

import (
	"net/http"
	"slices"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
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
		Filters: ParseFilters(ctx, query),
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
	mapping := map[string]arrow.Array{}
	for i := 0; i < int(resultRecord.NumCols()); i++ {
		mapping[resultRecord.ColumnName(i)] = resultRecord.Column(i)
	}
	var result []*v1.Value
	xc := &Compute{mapping: mapping}
	for _, metric := range metrics {
		value, err := xc.Metric(ctx, metric)
		if err != nil {
			logger.Get(ctx).Error("Failed calculating metric", "metric", metric)
			request.Internal(ctx, w)
			return
		}
		result = append(result, &v1.Value{
			Metric: metric,
			Value:  value,
		})
	}
	request.Write(ctx, w, &v1.Aggregate_Response{Results: result})
	return
}
