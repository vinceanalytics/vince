package timeseries

import (
	"errors"
	"fmt"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
)

const ShardWidth = 1 << 20

type batch struct {
	db        *pebble.DB
	translate *translation
	mutex     [models.TranslatedFieldsSize]map[uint64]*roaring.Bitmap
	bsi       [models.BSIFieldsSize]map[uint64]*roaring.Bitmap
	events    uint64
	id        uint64
	shard     uint64
	time      uint64
}

func newbatch(db *pebble.DB, tr *translation) *batch {
	b := &batch{translate: tr}
	for i := range b.mutex {
		b.mutex[i] = make(map[uint64]*roaring.Bitmap)
	}
	for i := range b.bsi {
		b.bsi[i] = make(map[uint64]*roaring.Bitmap)
	}
	b.id = tr.id
	b.shard = b.id / ShardWidth
	b.db = db
	return b
}

func (b *batch) save() error {
	if b.events == 0 {
		return nil
	}
	defer func() {
		b.translate.reset()
		for i := range b.mutex {
			clear(b.mutex[i])
		}
		for i := range b.bsi {
			clear(b.bsi[i])
		}
		b.events = 0
	}()
	ba := b.db.NewIndexedBatch()

	err := b.translate.flush(ba.Set)
	if err != nil {
		ba.Close()
		return err
	}
	err = b.flush(ba)
	if err != nil {
		ba.Close()
		return err
	}
	return ba.Commit(pebble.Sync)
}

func (b *batch) flush(ba *pebble.Batch) error {
	key := make([]byte, encoding.BitmapKeySize)
	sv := func(f models.Field, view uint64, bm *roaring.Bitmap) error {
		ts := time.UnixMilli(int64(view)).UTC()
		value := bm.ToBuffer()
		return errors.Join(
			ba.Merge(encoding.Bitmap(b.shard, 0, f, key), value, nil),
			ba.Merge(encoding.Bitmap(b.shard, view, f, key), value, nil),
			ba.Merge(encoding.Bitmap(b.shard, hour(ts), f, key), value, nil),
			ba.Merge(encoding.Bitmap(b.shard, day(ts), f, key), value, nil),
			ba.Merge(encoding.Bitmap(b.shard, week(ts), f, key), value, nil),
			ba.Merge(encoding.Bitmap(b.shard, month(ts), f, key), value, nil),
		)
	}
	for i := range b.mutex {
		f := models.Mutex(i)
		for view, bm := range b.mutex[i] {
			err := sv(f, view, bm)
			if err != nil {
				return fmt.Errorf("saving events bitmap %w", err)
			}
		}
	}
	for i := range b.bsi {
		f := models.BSI(i)
		for view, bm := range b.bsi[i] {
			err := sv(f, view, bm)
			if err != nil {
				return fmt.Errorf("saving events bitmap %w", err)
			}
		}
	}
	return nil
}

func (b *batch) add(m *models.Model) error {
	b.events++
	shard := (b.id + 1) / ShardWidth
	if shard != b.shard {
		err := b.save()
		if err != nil {
			return err
		}
		b.shard = shard
	}
	b.id = b.translate.Next()
	id := b.id
	ts := uint64(time.UnixMilli(m.Timestamp).Truncate(time.Minute).UnixMilli())
	b.time = ts
	if m.Timestamp > 0 {
		b.getBSI(models.Field_timestamp).BSI(id, m.Timestamp)
	}
	if m.Id != 0 {
		b.getBSI(models.Field_id).BSI(id, int64(m.Id))
	}
	if m.Bounce != 0 {
		b.getBSI(models.Field_bounce).BSI(id, int64(m.Bounce))
	}
	if m.Session {
		b.getMutex(models.Field_session).Bool(id, true)
	}
	if m.View {
		b.getMutex(models.Field_view).Bool(id, true)
	}
	if m.Duration > 0 {
		b.getBSI(models.Field_duration).BSI(id, m.Duration)
	}
	if m.City != 0 {
		b.getMutex(models.Field_city).Mutex(id, uint64(m.City))
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
	b.getMutex(field).Mutex(id, b.tr(field, value))
}

func (b *batch) getBSI(field models.Field) *roaring.Bitmap {
	idx := field.BSI()
	bs, ok := b.bsi[idx][b.time]
	if !ok {
		bs = roaring.NewBitmap()
		b.bsi[idx][b.time] = bs
	}
	return bs
}

func (b *batch) getMutex(field models.Field) *roaring.Bitmap {
	idx := field.Mutex()
	bs, ok := b.mutex[idx][b.time]
	if !ok {
		bs = roaring.NewBitmap()
		b.mutex[idx][b.time] = bs
	}
	return bs
}

func (b *batch) tr(field models.Field, value []byte) uint64 {
	return b.translate.Assign(field, value)
}

func hour(ts time.Time) uint64 {
	return uint64(compute.Hour(ts).UnixMilli())
}

func day(ts time.Time) uint64 {
	return uint64(compute.Date(ts).UnixMilli())
}

func week(ts time.Time) uint64 {
	return uint64(compute.Week(ts).UnixMilli())
}

func month(ts time.Time) uint64 {
	return uint64(compute.Month(ts).UnixMilli())
}
