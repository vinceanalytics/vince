package visitors

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

func Current(ctx context.Context, ts *timeseries.Timeseries, domain string) (visitors uint64, err error) {
	end := xtime.Now()
	start := end.Add(-5 * time.Minute)
	visitors = ts.Visitors(start, end, encoding.Minute, domain)
	return
}

func Visitors(ctx context.Context, ts *timeseries.Timeseries, domain string) (visitors uint64, err error) {
	end := xtime.Now()
	start := end.Add(-24 * time.Hour)
	visitors = ts.Visitors(start, end, encoding.Hour, domain)
	return
}
