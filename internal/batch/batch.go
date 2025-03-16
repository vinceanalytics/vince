package batch

import (
	"bytes"
	"iter"
	"sync/atomic"
	"time"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/translate"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

type Batch struct {
	tr    *translate.Transtate
	data  map[key]*roaring.Bitmap
	keys  [models.MutexFieldSize][][]byte
	id    uint64
	seq   *atomic.Uint64
	shard uint64
}

func New(tr *translate.Transtate, seq *atomic.Uint64) *Batch {
	return &Batch{
		tr:   tr,
		data: make(map[key]*roaring.Bitmap),
		seq:  seq,
	}
}

type key struct {
	field  models.Field
	shard  uint64
	exists bool
}

func (b *Batch) Reset() {
	for i := range b.keys {
		clear(b.keys[i])
		b.keys[i] = b.keys[i][:0]
	}
	clear(b.data)
}

func (b *Batch) IterKeys() iter.Seq2[models.Field, []byte] {
	return func(yield func(models.Field, []byte) bool) {
		for i := range b.keys {
			f := models.Mutex(i)
			for j := range b.keys[i] {
				if !yield(f, b.keys[i][j]) {
					return
				}
			}
		}
	}
}

type Container struct {
	Field     v1.Field
	Shard     uint64
	Existence bool
	Key       uint64
}

func (b *Batch) IterContainers() iter.Seq2[Container, *roaring.Container] {
	return func(yield func(Container, *roaring.Container) bool) {
		for k, v := range b.data {

			it, _ := v.Containers.Iterator(0)
			for it.Next() {
				key, co := it.Value()

				if !yield(Container{
					Field:     k.field,
					Shard:     k.shard,
					Existence: k.exists,
					Key:       key,
				}, co) {
					return
				}
			}
		}
	}
}

func (b *Batch) Next(ts time.Time, domain []byte) {
	b.id = b.seq.Add(1)
	b.shard = b.id / shardwidth.ShardWidth
	b.Int64(v1.Field_minute, compute.Minute(ts).UnixMilli())
	b.Int64(v1.Field_hour, compute.Hour(ts).UnixMilli())
	b.Int64(v1.Field_day, compute.Date(ts).UnixMilli())
	b.Int64(v1.Field_week, compute.Week(ts).UnixMilli())
	b.Int64(v1.Field_month, compute.Month(ts).UnixMilli())
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
	b.ra(f, false, func(ra *roaring.Bitmap) {
		ro2.WriteBSI(ra, b.id, value)
	})
}

func (b *Batch) Bool(f models.Field, value bool) {
	b.ra(f, false, func(ra *roaring.Bitmap) {
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
	b.ra(f, false, func(ra *roaring.Bitmap) {
		ro2.WriteMutex(ra, b.id, row)
	})
	b.ra(f, true, func(ra *roaring.Bitmap) {
		ra.DirectAdd(b.id % shardwidth.ShardWidth)
	})
}

func (b *Batch) ra(field models.Field, exists bool, f func(ra *roaring.Bitmap)) {
	k := key{
		field:  field,
		shard:  b.shard,
		exists: exists,
	}
	r := b.data[k]
	if r == nil {
		r = roaring.NewBitmap()
		b.data[k] = r
	}
	f(r)

}

func (b *Batch) translate(f models.Field, data []byte) uint64 {
	id, ok := b.tr.Get(f, data)
	if !ok {
		idx := models.AsMutex(f)
		b.keys[idx] = append(b.keys[idx], bytes.Clone(data))
	}
	return id
}
