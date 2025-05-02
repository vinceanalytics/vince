package batch

import (
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/storage/date"
	"github.com/vinceanalytics/vince/internal/storage/fields"
	"github.com/vinceanalytics/vince/internal/storage/translate/mapping"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

type Batch struct {
	tr    *mapping.Mapping
	data  map[key]*roaring.Bitmap
	id    uint64
	shard uint64
	ba    *pebble.Batch
}

func New(tr *mapping.Mapping, ba *pebble.Batch) *Batch {
	return &Batch{
		tr:   tr,
		data: make(map[key]*roaring.Bitmap),
		ba:   ba,
	}
}

type key struct {
	shard uint64
	field models.Field
	kind  v1.DataType
}

func (b *Batch) Next(ts time.Time, domain []byte) {
	b.id = b.tr.Next()
	b.shard = b.id / shardwidth.ShardWidth
	mins, hrs, dy, wk, mo, yy := date.Resolve(ts.UTC())
	b.Mutex(v1.Field_minute, mins)
	b.Mutex(v1.Field_hour, hrs)
	b.Mutex(v1.Field_day, dy)
	b.Mutex(v1.Field_week, wk)
	b.Mutex(v1.Field_month, mo)
	b.Mutex(v1.Field_year, yy)
}

func (b *Batch) Add(m *models.Model) error {
	if m.Timestamp == 0 || m.Id == 0 || len(m.Domain) == 0 {
		// Skip events without timestamp, id or domain
		return nil
	}
	b.Next(xtime.UnixMilli(m.Timestamp), m.Domain)

	b.Int64(v1.Field_timestamp, m.Timestamp)
	b.Int64(v1.Field_id, int64(m.Id))
	b.Bool(v1.Field_bounce, m.Bounce == 1)
	b.Bool(v1.Field_session, m.Session)
	b.Bool(v1.Field_view, m.View)
	b.Int64(v1.Field_duration, m.Duration)
	b.Mutex(v1.Field_city, uint64(m.City))

	err := b.String(v1.Field_domain, m.Domain)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_browser, m.Browser)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_browser_version, m.BrowserVersion)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_country, m.Country)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_device, m.Device)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_entry_page, m.EntryPage)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_event, m.Event)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_exit_page, m.ExitPage)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_host, m.Host)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_os, m.Os)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_os_version, m.OsVersion)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_page, m.Page)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_referrer, m.Referrer)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_source, m.Source)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_utm_campaign, m.UtmCampaign)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_utm_content, m.UtmContent)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_utm_medium, m.UtmMedium)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_utm_source, m.UtmSource)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_utm_term, m.UtmTerm)
	if err != nil {
		return err
	}
	err = b.String(v1.Field_subdivision1_code, m.Subdivision1Code)
	if err != nil {
		return err
	}
	return b.String(v1.Field_subdivision2_code, m.Subdivision2Code)
}

func (b *Batch) Int64(f models.Field, value int64) {
	if value == 0 {
		return
	}
	b.ra(f, v1.DataType_bsi, func(ra *roaring.Bitmap) {
		ro2.WriteBSI(ra, b.id, value)
	})
}

func (b *Batch) Bool(f models.Field, value bool) {
	b.ra(f, v1.DataType_bool, func(ra *roaring.Bitmap) {
		ro2.WriteBool(ra, b.id, value)
	})
}

func (b *Batch) String(f models.Field, value []byte) error {
	if len(value) == 0 {
		return nil
	}
	row, err := b.translate(f, value)
	if err != nil {
		return err
	}
	b.Mutex(f, row)
	return nil
}

func (b *Batch) Mutex(f models.Field, row uint64) {
	if row == 0 {
		return
	}
	b.ra(f, v1.DataType_mutex, func(ra *roaring.Bitmap) {
		ro2.WriteMutex(ra, b.id, row)
	})
	b.ra(f, v1.DataType_exists, func(ra *roaring.Bitmap) {
		ra.DirectAdd(b.id % shardwidth.ShardWidth)
	})
}

func (b *Batch) Close() error {
	return b.ba.Close()
}

func (b *Batch) Commit() error {

	var data fields.DataKey

	for k, v := range b.data {

		it, _ := v.Containers.Iterator(0)
		for it.Next() {
			key, co := it.Value()

			data.Make(k.field, k.kind, k.shard, key)
			err := b.ba.Merge(data[:], co.Encode(), nil)
			if err != nil {
				return err
			}

		}
	}
	return b.ba.Commit(nil)
}

func (b *Batch) ra(field models.Field, kind v1.DataType, f func(ra *roaring.Bitmap)) {
	k := key{
		field: field,
		shard: b.shard,
		kind:  kind,
	}
	r := b.data[k]
	if r == nil {
		r = roaring.NewBitmap()
		b.data[k] = r
	}
	f(r)

}

func (b *Batch) translate(f models.Field, data []byte) (uint64, error) {
	id, ok := b.tr.GetOrCreate(f, data)
	if !ok {
		tk := fields.MakeTranslationKey(f, b.shard, data)
		ti := fields.MakeTranslationID(f, b.shard, id)
		err := b.ba.Set(tk, ti[len(ti)-8:], nil)
		if err != nil {
			return 0, err
		}
		err = b.ba.Set(ti, data, nil)
		if err != nil {
			return 0, err
		}

	}
	return id, nil
}
