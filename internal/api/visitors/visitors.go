package visitors

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func Current(ctx context.Context, ts *timeseries.Timeseries, domain string) (visitors uint64, err error) {
	end := time.Now().UTC()
	start := end.Add(-5 * time.Minute)
	r := roaring.NewBitmap()
	err = query.Minute.Range(start, end, func(t time.Time) error {
		view := uint64(t.UnixMilli())
		for shard := range ts.Shards(ctx) {
			match := ts.Domain(ctx, shard, view, domain)
			if match.IsEmpty() {
				continue
			}
			ts.NewBitmap(ctx, shard, view, models.Field_id).
				ExtractBSI(shard, match, func(_ uint64, value int64) {
					r.Set(uint64(value))
				})
		}
		return nil
	})
	visitors = uint64(r.GetCardinality())
	return
}

func Visitors(ctx context.Context, ts *timeseries.Timeseries, domain string) (visitors uint64, err error) {
	b := roaring.NewBitmap()
	for shard := range ts.Shards(ctx) {
		match := ts.Domain(ctx, shard, 0, domain)
		if match.IsEmpty() {
			continue
		}
		ts.NewBitmap(ctx, shard, 0, models.Field_id).
			ExtractBSI(shard, match, func(_ uint64, value int64) {
				b.Set(uint64(value))
			})
	}
	visitors = uint64(b.GetCardinality())
	return
}
