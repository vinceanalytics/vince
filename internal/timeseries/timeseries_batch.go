package timeseries

import (
	"errors"
	"fmt"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/shards"
	"github.com/vinceanalytics/vince/internal/util/oracle"
	"github.com/vinceanalytics/vince/internal/util/xtime"
)

const ShardWidth = 1 << 20

type batch struct {
	db             *shards.DB
	translate      *translation
	mutex          [models.MutexFieldSize]*ro2.Bitmap
	mutexExistence [models.MutexFieldSize]*ro2.Bitmap
	bsi            [models.BSIFieldsSize]*ro2.Bitmap
	views          shards.Views
	events         uint64
	id             uint64
	shard          uint64
	time           uint64
	key            encoding.Key
}

func newbatch(db *shards.DB, tr *translation) *batch {
	oracle.Records.Store(tr.id)
	b := &batch{
		translate: tr,
		id:        tr.id,
		shard:     tr.id / ShardWidth,
		db:        db,
	}
	for i := range b.mutex {
		b.mutex[i] = ro2.NewBitmap()
		b.mutexExistence[i] = ro2.NewBitmap()
	}
	for i := range b.bsi {
		b.bsi[i] = ro2.NewBitmap()
	}
	b.views.Init()
	return b
}

// saves only current timestamp.
func (b *batch) save() error {
	if b.events == 0 {
		return nil
	}
	defer func() {
		b.translate.reset()
		for i := range b.mutex {
			b.mutex[i].Containers.Reset()
			b.mutexExistence[i].Containers.Reset()
		}
		for i := range b.bsi {
			b.bsi[i].Containers.Reset()
		}
		b.events = 0
		oracle.Records.Store(b.id)
		b.views.Reset()
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

	for i := range b.mutex {
		f := models.Mutex(i)
		bm := b.mutex[i]
		if !bm.Any() {
			continue
		}
		err := errors.Join(
			b.mergeData(ba, f, bm),
			b.mergeExists(ba, f, b.mutexExistence[i]),
		)
		if err != nil {
			return fmt.Errorf("saving events bitmap %w", err)
		}
	}
	for i := range b.bsi {
		f := models.BSI(i)
		bm := b.bsi[i]
		if !bm.Any() {
			continue
		}
		err := b.mergeData(ba, f, bm)
		if err != nil {
			return fmt.Errorf("saving events bitmap %w", err)
		}
	}
	return nil
}

func (b *batch) mergeData(ba *pebble.Batch, field models.Field, bm *ro2.Bitmap) error {
	return b.merge(ba, field, bm, b.data)
}

func (b *batch) mergeExists(ba *pebble.Batch, field models.Field, bm *ro2.Bitmap) error {
	return b.merge(ba, field, bm, b.exists)
}

func (b *batch) merge(ba *pebble.Batch, field models.Field, bm *ro2.Bitmap, enc func(field models.Field, co uint64) []byte) error {
	ci, _ := bm.Containers.Iterator(0)
	for ci.Next() {
		key, co := ci.Value()
		value := ro2.EncodeContainer(co)
		err := ba.Merge(enc(field, key), value, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *batch) data(field models.Field, co uint64) []byte {
	b.key.WriteData(field, co)
	return b.key.Bytes()
}

func (b *batch) exists(field models.Field, co uint64) []byte {
	b.key.WriteExistence(field, co)
	return b.key.Bytes()
}

func (b *batch) add(m *models.Model) error {
	if m.Timestamp == 0 || m.Id == 0 {
		// Skip events without timestamp, id
		return nil
	}
	ts := uint64(xtime.UnixMilli(m.Timestamp).Truncate(time.Minute).UnixMilli())
	if ts != b.time {
		err := b.save()
		if err != nil {
			return err
		}
		b.time = ts
	}
	shard := (b.id + 1) / ShardWidth
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
	ro2.WriteBSI(b.getBSI(models.Field_timestamp), id, m.Timestamp)
	ro2.WriteBSI(b.getBSI(models.Field_id), id, int64(m.Id))
	if m.Bounce != 0 {
		ro2.WriteBool(b.getBSI(models.Field_bounce), id, m.Bounce == 1)
	}
	if m.Session {
		ro2.WriteBool(b.mutex[models.Field_session.Mutex()], id, true)
	}
	if m.View {
		ro2.WriteBool(b.mutex[models.Field_view.Mutex()], id, true)
	}
	if m.Duration > 0 {
		ro2.WriteBSI(b.getBSI(models.Field_duration), id, m.Duration)
	}
	if m.City != 0 {
		ro2.WriteMutex(b.mutex[models.Field_city.Mutex()], id, uint64(m.City))
	}
	b.set(models.Field_browser, id, m.Browser)
	b.set(models.Field_browser_version, id, m.BrowserVersion)
	b.set(models.Field_country, id, m.Country)
	b.set(models.Field_device, id, m.Device)
	b.set(models.Field_domain, id, m.Domain)
	b.set(models.Field_entry_page, id, m.EntryPage)
	b.set(models.Field_event, id, m.Event)
	b.set(models.Field_exit_page, id, m.ExitPage)
	b.set(models.Field_host, id, m.Host)
	b.set(models.Field_os, id, m.Os)
	b.set(models.Field_os_version, id, m.OsVersion)
	b.set(models.Field_page, id, m.Page)
	b.set(models.Field_referrer, id, m.Referrer)
	b.set(models.Field_source, id, m.Source)
	b.set(models.Field_utm_campaign, id, m.UtmCampaign)
	b.set(models.Field_utm_content, id, m.UtmContent)
	b.set(models.Field_utm_medium, id, m.UtmMedium)
	b.set(models.Field_utm_source, id, m.UtmSource)
	b.set(models.Field_utm_term, id, m.UtmTerm)
	b.set(models.Field_subdivision1_code, id, m.Subdivision1Code)
	b.set(models.Field_subdivision2_code, id, m.Subdivision2Code)
	return nil
}

func (b *batch) set(field models.Field, id uint64, value []byte) {
	if len(value) == 0 {
		return
	}
	idx := field.Mutex()
	ro2.WriteMutex(b.mutex[idx], id, b.tr(field, value))
	b.mutexExistence[idx].DirectAdd(id % ShardWidth)
}

func (b *batch) getBSI(field models.Field) *ro2.Bitmap {
	idx := field.BSI()
	bs := b.bsi[idx]
	if bs != nil {
		return bs
	}
	b.bsi[idx] = ro2.NewBitmap()
	return b.bsi[idx]
}

func (b *batch) tr(field models.Field, value []byte) uint64 {
	return b.translate.Assign(field, value)
}
