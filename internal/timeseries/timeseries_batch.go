package timeseries

import (
	"encoding/binary"

	"github.com/cockroachdb/pebble"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/shards"
	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

type Key struct {
	View       uint64
	Domain     uint64
	Resolution encoding.Resolution
	Field      models.Field
	Existence  bool
}

func (k *Key) Encode(co uint64, b []byte) []byte {
	if k.Existence {
		b = append(b[:0], keys.DataExistsPrefix...)
	} else {
		b = append(b[:0], keys.DataPrefix...)
	}
	b = append(b, byte(k.Resolution))
	b = binary.BigEndian.AppendUint64(b, k.View)
	b = binary.BigEndian.AppendUint64(b, k.Domain)
	b = append(b, byte(k.Field))
	b = binary.BigEndian.AppendUint64(b, co)
	return b
}

type batch struct {
	db        *shards.DB
	translate *translation
	all       map[Key]*ro2.Bitmap
	key       []byte
	views     [encoding.Month + 1]uint64
	events    uint64
	id        uint64
	shard     uint64

	domainId uint64
}

func newbatch(db *shards.DB, tr *translation) *batch {
	oracle.Records.Store(tr.id)
	b := &batch{
		translate: tr,
		id:        tr.id,
		shard:     tr.id / shardwidth.ShardWidth,
		db:        db,
		all:       make(map[Key]*roaring.Bitmap),
	}
	return b
}

func (b *batch) setTs(timestamp int64) {
	ts := xtime.UnixMilli(timestamp)
	b.views[encoding.Minute] = uint64(compute.Minute(ts).UnixMilli())
	b.views[encoding.Hour] = uint64(compute.Hour(ts).UnixMilli())
	b.views[encoding.Week] = uint64(compute.Week(ts).UnixMilli())
	b.views[encoding.Day] = uint64(compute.Date(ts).UnixMilli())
}

// saves only current timestamp.
func (b *batch) save() error {
	if b.events == 0 {
		return nil
	}
	defer func() {
		b.translate.reset()
		clear(b.all)
		b.events = 0
		oracle.Records.Store(b.id)
	}()
	oba := b.db.Get().NewBatch()

	err := b.translate.flush(oba.Set)
	if err != nil {
		oba.Close()
		return err
	}
	err = oba.Commit(pebble.Sync)
	if err != nil {
		return err
	}
	sh := b.db.Shard(b.shard)
	ba := sh.DB.NewBatch()
	err = b.flush(ba)
	if err != nil {
		ba.Close()
		return err
	}
	return ba.Commit(pebble.Sync)
}

func (b *batch) flush(ba *pebble.Batch) error {
	for k, bm := range b.all {
		err := b.merge(ba, k, bm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *batch) merge(ba *pebble.Batch, ke Key, bm *ro2.Bitmap) error {
	ci, _ := bm.Containers.Iterator(0)
	for ci.Next() {
		key, co := ci.Value()
		value := ro2.EncodeContainer(co)
		b.key = ke.Encode(key, b.key)
		err := ba.Merge(b.key, value, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *batch) add(m *models.Model) error {
	if m.Timestamp == 0 || m.Id == 0 {
		// Skip events without timestamp, id
		return nil
	}
	shard := (b.id + 1) / shardwidth.ShardWidth
	if shard != b.shard {
		err := b.save()
		if err != nil {
			return err
		}
		b.shard = shard
	}
	b.events++

	b.id = b.translate.Next()
	id := b.id
	b.setTs(m.Timestamp)
	b.bs(v1.Field_id, id, int64(m.Id))
	if m.Bounce != 0 {
		b.boolean(v1.Field_bounce, id, m.Bounce == 1)
	}
	if m.Session {
		b.boolean(v1.Field_session, id, true)
	}
	if m.View {
		b.boolean(v1.Field_view, id, true)
	}
	if m.Duration > 0 {
		b.bs(v1.Field_duration, id, m.Duration)
	}
	if m.City != 0 {
		b.mx(v1.Field_city, id, uint64(m.City))
	}
	b.set(v1.Field_browser, id, m.Browser)
	b.set(v1.Field_browser_version, id, m.BrowserVersion)
	b.set(v1.Field_country, id, m.Country)
	b.set(v1.Field_device, id, m.Device)

	// domain is stored as part of the key, we only save existence bit
	b.domainId = b.tr(v1.Field_domain, m.Domain)
	b.mxExixtenceOnly(v1.Field_domain, id)

	b.set(v1.Field_entry_page, id, m.EntryPage)
	b.set(v1.Field_event, id, m.Event)
	b.set(v1.Field_exit_page, id, m.ExitPage)
	b.set(v1.Field_host, id, m.Host)
	b.set(v1.Field_os, id, m.Os)
	b.set(v1.Field_os_version, id, m.OsVersion)
	b.set(v1.Field_page, id, m.Page)
	b.set(v1.Field_referrer, id, m.Referrer)
	b.set(v1.Field_source, id, m.Source)
	b.set(v1.Field_utm_campaign, id, m.UtmCampaign)
	b.set(v1.Field_utm_content, id, m.UtmContent)
	b.set(v1.Field_utm_medium, id, m.UtmMedium)
	b.set(v1.Field_utm_source, id, m.UtmSource)
	b.set(v1.Field_utm_term, id, m.UtmTerm)
	b.set(v1.Field_subdivision1_code, id, m.Subdivision1Code)
	b.set(v1.Field_subdivision2_code, id, m.Subdivision2Code)
	return nil
}

func (b *batch) bs(field models.Field, id uint64, value int64) {
	for i := range b.views {
		ro2.WriteBSI(b.ra(Key{
			Resolution: encoding.Resolution(i),
			Field:      field,
			View:       b.views[i],
			Domain:     b.domainId,
		}), id, value)
	}
}

func (b *batch) boolean(field models.Field, id uint64, value bool) {
	for i := range b.views {
		ro2.WriteBool(b.ra(Key{
			Resolution: encoding.Resolution(i),
			Field:      field,
			View:       b.views[i],
			Domain:     b.domainId,
		}), id, value)
	}
}

func (b *batch) set(field models.Field, id uint64, value []byte) {
	if len(value) == 0 {
		return
	}
	b.mx(field, id, b.tr(field, value))

}

func (b *batch) mx(field models.Field, id uint64, value uint64) {

	for i := range b.views {
		ro2.WriteMutex(b.ra(Key{
			Resolution: encoding.Resolution(i),
			Field:      field,
			View:       b.views[i],
			Domain:     b.domainId,
		}), id, value)
		b.ra(Key{
			Resolution: encoding.Resolution(i),
			Field:      field,
			View:       b.views[i],
			Domain:     b.domainId,
			Existence:  true,
		}).DirectAdd(id % shardwidth.ShardWidth)
	}
}

func (b *batch) mxExixtenceOnly(field models.Field, id uint64) {
	for i := range b.views {
		b.ra(Key{
			Resolution: encoding.Resolution(i),
			Field:      field,
			View:       b.views[i],
			Domain:     b.domainId,
			Existence:  true,
		}).DirectAdd(id % shardwidth.ShardWidth)
	}
}

func (b *batch) ra(key Key) *ro2.Bitmap {
	if a, ok := b.all[key]; ok {
		return a
	}
	a := ro2.NewBitmap()
	b.all[key] = a
	return a
}

func (b *batch) tr(field models.Field, value []byte) uint64 {
	return b.translate.Assign(field, value)
}
