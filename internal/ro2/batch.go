package ro2

import (
	"encoding/binary"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
)

const ShardWidth = 1 << 20

func (b *Store) Add(m *models.Model) error {
	shard := (b.id + 1) / ShardWidth
	if shard != b.shard {
		// we havs changed shards. Persist the current batch before continuing
		b.shard = shard
	}
	b.id++
	id := b.id
	ts := uint64(time.UnixMilli(m.Timestamp).Truncate(time.Minute).UnixMilli())
	b.key.Time = ts
	if m.Timestamp != 0 {
		b.get(models.Field_timestamp).SetValue(id, m.Timestamp)
	}
	if m.Id != 0 {
		b.get(models.Field_id).SetValue(id, int64(m.Id))
	}
	if m.Bounce != 0 {
		b.get(models.Field_bounce).SetValue(id, int64(m.Bounce))
	}
	if m.Session {
		b.get(models.Field_session).SetValue(id, 1)
	}
	if m.View {
		b.get(models.Field_view).SetValue(id, 1)
	}
	if m.Duration != 0 {
		b.get(models.Field_duration).SetValue(id, m.Duration)
	}
	if m.Duration != 0 {
		b.get(models.Field_city).SetValue(id, int64(m.City))
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

func (b *Store) set(field models.Field, id uint64, value []byte) {
	if len(value) == 0 {
		return
	}
	b.get(field).SetValue(id, b.tr(field, value))
}

func (b *Store) get(field models.Field) *roaring.BSI {
	b.key.Field = field
	bs, ok := b.data[b.key]
	if !ok {
		bs = roaring.NewDefaultBSI()
		b.data[b.key] = bs
	}
	return bs
}

func (b *Store) tr(field models.Field, value []byte) int64 {
	return int64(b.AssignUid(field, value))
}

func (b *Store) SaveBatch() (err error) {
	if len(b.data) == 0 {
		return
	}
	defer func() {
		clear(b.data)
		b.enc.Reset()
	}()

	err = b.Flush()
	if err != nil {
		return
	}

	tx := b.db.NewTransaction(true)
	defer func() {
		if err != nil {
			tx.Discard()
		} else {
			err = tx.Commit()
		}
	}()

	// start by saving the current record id, even if some ops failed we will not
	//  mess up search
	seq := b.enc.Allocate(8)
	binary.BigEndian.PutUint64(seq, b.id)

	err = tx.Set(b.enc.TranslateSeq(models.Field_unknown), seq)
	if err != nil {
		return err
	}

	for k, v := range b.data {
		err = b.saveTs(tx, k, v)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				err = tx.Commit()
				if err != nil {
					return
				}
				tx = b.db.NewTransaction(true)
				b.txnCount++
				err = b.saveTs(tx, k, v)
				if err != nil {
					return
				}
				continue
			}
			return err
		}
	}
	return nil
}

func (b *Store) saveTs(tx *badger.Txn, key encoding.Key, value *roaring.BSI) error {
	ts := time.UnixMilli(int64(key.Time)).UTC()
	return value.Each(func(idx byte, bs *roaring.Bitmap) error {
		return errors.Join(
			b.save(tx, b.enc.Bitmap(0, b.shard, key.Field, byte(idx)), bs),        // global
			b.save(tx, b.enc.Bitmap(key.Time, b.shard, key.Field, byte(idx)), bs), //minute
			b.save(tx, b.enc.Bitmap(hour(ts), b.shard, key.Field, byte(idx)), bs),
			b.save(tx, b.enc.Bitmap(day(ts), b.shard, key.Field, byte(idx)), bs),
			b.save(tx, b.enc.Bitmap(week(ts), b.shard, key.Field, byte(idx)), bs),
			b.save(tx, b.enc.Bitmap(month(ts), b.shard, key.Field, byte(idx)), bs),
		)
	})
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

func (b *Store) save(tx *badger.Txn, key []byte, value *roaring.Bitmap) error {
	it, err := tx.Get(key)
	var data []byte
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		data = value.ToBuffer()
	} else {
		err = it.Value(func(val []byte) error {
			bs := roaring.Or(roaring.FromBuffer(val), value)
			data = bs.ToBuffer()
			return err
		})
		if err != nil {
			return err
		}
	}
	return tx.Set(key, data)
}
