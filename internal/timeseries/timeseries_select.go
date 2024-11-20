package timeseries

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/timeseries/cursor"
	"github.com/vinceanalytics/vince/internal/util/data"
	"github.com/vinceanalytics/vince/internal/web/query"
)

type ScanCall func(cu *cursor.Cursor, field models.Field, view, shard uint64, columns *ro2.Bitmap) error

func (ts *Timeseries) Select(
	ctx context.Context,
	values models.BitSet,
	domain string, start,
	end time.Time,
	intrerval query.Interval,
	filters query.Filters,
	cb ScanCall) error {
	m := ts.compile(filters)
	domainId := ts.Translate(models.Field_domain, []byte(domain))
	return ts.Scan(domainId, encoding.Resolution(intrerval), start,
		end, m, values, cb)
}

func (ts *Timeseries) Find(ctx context.Context, field models.Field, id uint64) (value string) {
	data.Get(ts.db.Get(), encoding.TranslateID(field, id), func(val []byte) error {
		value = string(val)
		return nil
	})
	return
}
