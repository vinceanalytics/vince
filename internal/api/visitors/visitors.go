package visitors

import (
	"context"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

func Current(ctx context.Context, ts *timeseries.Timeseries, domain string) (visitors uint64, err error) {
	visitors = ts.Realtime(domain)
	return
}

func Visitors(ctx context.Context, ts *timeseries.Timeseries, domain string) (visitors uint64, err error) {
	b := roaring.NewBitmap()
	ts.ScanGlobal(models.Field_id, domain, func(shard uint64, columns, ra *roaring.Bitmap) {
		ra.ExtractBSI(shard, columns, func(id uint64, value int64) {
			b.Set(uint64(value))
		})
	})
	visitors = uint64(b.GetCardinality())
	return
}
