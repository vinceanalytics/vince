package stats

import (
	"net/http"
	"time"

	"github.com/apache/arrow/go/v15/arrow/compute"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/request"
	"github.com/vinceanalytics/vince/session"
)

func Realtime(w http.ResponseWriter, r *http.Request) {
	var req v1.Realtime_Request
	req.SiteId = r.URL.Query().Get("site_id")
	ctx := r.Context()
	if err := request.Get(r.Context()).Validate(&req); err != nil {
		logger.Get(ctx).Error("Failed validating request body", "err", err)
		request.Error(ctx, w, http.StatusBadRequest, err.Error())
		return
	}
	now := time.Now().UTC()
	firstTime := now.Add(-5 * time.Minute)
	result, err := session.Get(ctx).Scan(ctx,
		firstTime.UnixMilli(),
		now.UnixMilli(),
		&v1.Filters{
			Projection: []v1.Filters_Projection{
				v1.Filters_ID,
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
	res, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(result.Column(0)))
	if err != nil {
		logger.Get(ctx).Error("Failed computing unique user id", "err", err)
		request.Internal(ctx, w)
		return
	}
	defer res.Release()
	request.Write(ctx, w, &v1.Realtime_Response{Visitors: uint64(res.Len())})
}
