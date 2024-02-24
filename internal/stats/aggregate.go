package stats

import (
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/request"
	"github.com/vinceanalytics/vince/internal/session"
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
	res, err := compute.Aggregate(ctx, session.Get(ctx), &req)
	if err != nil {
		logger.Get(ctx).Error("Failed scanning", "err", err)
		request.Internal(ctx, w)
		return
	}
	request.Write(ctx, w, res)
}
