package stats

import (
	"context"
	"time"

	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/vinceanalytics/staples/staples/db"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/logger"
)

func CurrentVisitors(ctx context.Context, domain string, scanner db.Scanner) int64 {
	log := logger.Get(ctx).With("call", "stats.CurrentVisitors", "domain", domain)
	now := time.Now().UTC()
	firstTime := now.Add(-5 * time.Minute)
	r, err := scanner.Scan(ctx,
		firstTime.UnixMilli(),
		now.UnixMilli(),
		&v1.Filters{
			Projection: []v1.Filters_Projection{
				v1.Filters_ID,
			},
			List: []*v1.Filter{
				{Column: v1.Filter_Domain, Op: v1.Filter_equal, Value: domain},
			},
		},
	)
	if err != nil {
		log.Error("Failed scanning", "err", err)
		return 0
	}
	defer r.Release()
	res, err := compute.Unique(ctx, compute.NewDatumWithoutOwning(r.Column(0)))
	if err != nil {
		log.Error("Failed computing unique user id", "err", err)
		return 0
	}
	defer res.Release()
	return res.Len()
}
