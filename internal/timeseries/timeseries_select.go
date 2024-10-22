package timeseries

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (ts *Timeseries) Select(ctx context.Context, domain string, start,
	end time.Time, intrerval query.Interval, filters query.Filters, cb func(shard, view uint64, columns *roaring.Bitmap) error) error {
	m := ts.compile(filters)
	return intrerval.Range(start, end, func(t time.Time) error {
		view := uint64(t.UnixMilli())
		for shard := range oracle.Shards() {
			match := ts.Domain(ctx, shard, view, domain)
			if match.IsEmpty() {
				return nil
			}
			columns := m.Apply(ctx, ts, shard, view, match)
			if columns.IsEmpty() {
				return nil
			}
			err := cb(shard, view, columns)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (ts *Timeseries) Domain(ctx context.Context, shard, view uint64, name string) *roaring.Bitmap {
	bs := ts.NewBitmap(ctx, shard, view, models.Field_domain)
	return bs.Row(shard, ts.Translate(models.Field_domain, []byte(name)))
}

func (ts *Timeseries) Find(ctx context.Context, field models.Field, id uint64) (value string) {
	data.Get(ts.db, encoding.TranslateID(field, id), func(val []byte) error {
		value = string(val)
		return nil
	})
	return
}
