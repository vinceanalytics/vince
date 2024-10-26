package timeseries

import (
	"context"
	"time"

	"github.com/bits-and-blooms/bitset"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (ts *Timeseries) Select(
	ctx context.Context,
	values *bitset.BitSet,
	domain string, start,
	end time.Time,
	intrerval query.Interval,
	filters query.Filters,
	cb func(shard, view uint64, columns *roaring.Bitmap, data FieldsData)) {
	m := ts.compile(filters)
	m.Set(true, models.Field_domain, ts.Translate(models.Field_domain, []byte(domain)))
	views := ts.Shards(intrerval.Range(start, end))
	data := ts.Scan(views, m, values)
	for shard := range data {
		sx := &data[shard]
		if len(sx.Views) == 0 {
			continue
		}
		for view, data := range sx.Views {
			cb(uint64(shard), view, sx.Columns, *data)
		}
	}
	return
}

func (ts *Timeseries) Find(ctx context.Context, field models.Field, id uint64) (value string) {
	data.Get(ts.db, encoding.TranslateID(field, id), func(val []byte) error {
		value = string(val)
		return nil
	})
	return
}
