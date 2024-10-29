package timeseries

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/data"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func (ts *Timeseries) Select(
	ctx context.Context,
	values models.BitSet,
	domain string, start,
	end time.Time,
	intrerval query.Interval,
	filters query.Filters,
	cb func(field models.Field, view, shard uint64, columns, ra *roaring.Bitmap) bool) {
	m := ts.compile(filters)
	m.Set(true, models.Field_domain, ts.Translate(models.Field_domain, []byte(domain)))
	ts.Scan(encoding.Resolution(intrerval), start,
		end, m, values, cb)
	return
}

func (ts *Timeseries) Find(ctx context.Context, field models.Field, id uint64) (value string) {
	data.Get(ts.db, encoding.TranslateID(field, id), func(val []byte) error {
		value = string(val)
		return nil
	})
	return
}
