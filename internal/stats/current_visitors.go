package stats

import (
	"net/http"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/logger"
	"github.com/vinceanalytics/vince/internal/request"
	"github.com/vinceanalytics/vince/internal/session"
	"github.com/vinceanalytics/vince/internal/tenant"
)

func Realtime(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	req := v1.Realtime_Request{
		SiteId: query.Get("site_id"),
	}
	if !request.Validate(ctx, w, &req) {
		return
	}
	now := time.Now().UTC()
	firstTime := now.Add(-5 * time.Minute)
	result, err := session.Get(ctx).Scan(ctx,
		tenant.Get(ctx),
		firstTime.UnixMilli(),
		now.UnixMilli(),
		&v1.Filters{
			Projection: []v1.Filters_Projection{
				v1.Filters_id,
			},
			List: []*v1.Filter{
				{Property: v1.Property_domain, Op: v1.Filter_equal, Value: req.SiteId},
			},
		},
	)
	if err != nil {
		logger.Get(ctx).Error("Failed scanning", "err", err)
		request.Internal(ctx, w)
		return
	}
	defer result.Release()
	m := compute.NewCompute(result)
	visitors, err := m.Visitors(ctx)
	if err != nil {
		logger.Get(ctx).Error("Failed computing unique user id", "err", err)
		request.Internal(ctx, w)
		return
	}
	request.Write(ctx, w, uint64(visitors))
}
