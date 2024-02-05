package stats

import (
	"context"
	"time"

	"github.com/apache/arrow/go/v15/arrow/compute"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/logger"
	"github.com/vinceanalytics/vince/session"
)

func Realtime(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	now := time.Now().UTC()
	firstTime := now.Add(-5 * time.Minute)
	r, err := session.Get(ctx).Scan(ctx,
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
		return nil, InternalError
	}
	defer r.Release()
	res, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(r.Column(0)))
	if err != nil {
		logger.Get(ctx).Error("Failed computing unique user id", "err", err)
		return nil, InternalError
	}
	defer res.Release()
	return &v1.Realtime_Response{Visitors: uint64(res.Len())}, nil
}
