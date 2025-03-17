package batch

import (
	"bytes"
	"encoding/binary"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/storage/date"
	"github.com/vinceanalytics/vince/internal/storage/fields"
	"github.com/vinceanalytics/vince/internal/storage/translate"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

type Batch struct {
	tr     *translate.Transtate
	data   map[key]*roaring.Bitmap
	seq    *atomic.Uint64
	keys   [models.MutexFieldSize][][]byte
	ids    [models.MutexFieldSize][]uint64
	shards [models.MutexFieldSize][]uint64
	id     uint64
	shard  uint64
}

func New(tr *translate.Transtate, seq *atomic.Uint64) *Batch {
	return &Batch{
		tr:   tr,
		data: make(map[key]*roaring.Bitmap),
		seq:  seq,
	}
}

type key struct {
	shard uint64
	field models.Field
	kind  v1.DataType
}

func (b *Batch) Reset() {
	for i := range b.keys {
		clear(b.keys[i])
		clear(b.ids[i])
		b.keys[i] = b.keys[i][:0]
	}
	clear(b.data)
}

func (b *Batch) Next(ts time.Time, domain []byte) {
	b.id = b.seq.Add(1)
	b.shard = b.id / shardwidth.ShardWidth
	mins, hrs, dy, wk, mo, yy := date.Resolve(ts.UTC())
	b.Mutex(v1.Field_minute, mins)
	b.Mutex(v1.Field_hour, hrs)
	b.Mutex(v1.Field_day, dy)
	b.Mutex(v1.Field_week, wk)
	b.Mutex(v1.Field_month, mo)
	b.Mutex(v1.Field_year, yy)
}

func (b *Batch) Add(m *models.Model) {
	if m.Timestamp == 0 || m.Id == 0 || len(m.Domain) == 0 {
		// Skip events without timestamp, id or domain
		return
	}
	b.Next(xtime.UnixMilli(m.Timestamp), m.Domain)

	b.Int64(v1.Field_timestamp, m.Timestamp)
	b.Int64(v1.Field_id, int64(m.Id))
	b.Bool(v1.Field_bounce, m.Bounce == 1)
	b.Bool(v1.Field_session, m.Session)
	b.Bool(v1.Field_view, m.View)
	b.Int64(v1.Field_duration, m.Duration)
	b.Mutex(v1.Field_city, uint64(m.City))

	b.String(v1.Field_domain, m.Domain)
	b.String(v1.Field_browser, m.Browser)
	b.String(v1.Field_browser_version, m.BrowserVersion)
	b.String(v1.Field_country, m.Country)
	b.String(v1.Field_device, m.Device)
	b.String(v1.Field_entry_page, m.EntryPage)
	b.String(v1.Field_event, m.Event)
	b.String(v1.Field_exit_page, m.ExitPage)
	b.String(v1.Field_host, m.Host)
	b.String(v1.Field_os, m.Os)
	b.String(v1.Field_os_version, m.OsVersion)
	b.String(v1.Field_page, m.Page)
	b.String(v1.Field_referrer, m.Referrer)
	b.String(v1.Field_source, m.Source)
	b.String(v1.Field_utm_campaign, m.UtmCampaign)
	b.String(v1.Field_utm_content, m.UtmContent)
	b.String(v1.Field_utm_medium, m.UtmMedium)
	b.String(v1.Field_utm_source, m.UtmSource)
	b.String(v1.Field_utm_term, m.UtmTerm)
	b.String(v1.Field_subdivision1_code, m.Subdivision1Code)
	b.String(v1.Field_subdivision2_code, m.Subdivision2Code)
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

func (b *Batch) String(f models.Field, value []byte) {
	if len(value) == 0 {
		return
	}
	row := b.translate(f, value)
	b.Mutex(f, row)
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

func (b *Batch) Apply(wba *pebble.Batch) error {
	defer b.Reset()

	trKey := fields.MakeTranslationKey(0, 0, nil)
	trID := fields.MakeTranslationID(0, 0, 0)

	for i := range b.keys {
		f := models.Mutex(i)

		for j := range b.keys[i] {
			key := b.keys[i][j]
			id := b.ids[i][j]
			shard := b.shards[i][j]
			trKey[fields.FieldOffset] = byte(f)
			binary.BigEndian.PutUint64(trKey[fields.TranslationShardOffset:], shard)
			trKey = append(trKey[:fields.TranslationKeyOffset], key...)
			trID[fields.FieldOffset] = byte(f)
			binary.BigEndian.PutUint64(trID[fields.TranslationShardOffset:], shard)
			binary.BigEndian.PutUint64(trID[fields.TranslationIDOffset:], id)

			err := wba.Set(trKey, trID[fields.TranslationIDOffset:], nil)
			if err != nil {
				return err
			}
			err = wba.Set(trID, trKey[fields.TranslationKeyOffset:], nil)
			if err != nil {
				return err
			}
		}
	}

	var data fields.DataKey

	for k, v := range b.data {

		it, _ := v.Containers.Iterator(0)
		for it.Next() {
			key, co := it.Value()

			data.Make(k.field, k.kind, k.shard, key)
			err := wba.Merge(data[:], co.Encode(), nil)
			if err != nil {
				return err
			}

		}
	}

	return nil
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

func (b *Batch) translate(f models.Field, data []byte) uint64 {
	id, ok := b.tr.Get(f, b.shard, data)
	if !ok {
		idx := models.AsMutex(f)
		b.keys[idx] = append(b.keys[idx], bytes.Clone(data))
		b.ids[idx] = append(b.ids[idx], id)
		b.shards[idx] = append(b.shards[idx], b.shard)
	}
	return id
}
