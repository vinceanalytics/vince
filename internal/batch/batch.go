package batch

import (
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
)

type KV interface {
	RecordID() uint64
	Translate(field v1.Field, value []byte) uint64
}

type Batch struct {
	data    map[encoding.Key]*roaring.BSI
	domains map[uint32]uint64
	offsets []uint32
	key     encoding.Key
	enc     encoding.Encoding
}

func NewBatch() *Batch {
	b := &Batch{
		data:    make(map[encoding.Key]*roaring.BSI),
		domains: make(map[uint32]uint64),
		offsets: make([]uint32, 0, 65),
	}
	// a lot of small allocations happens during batching. We pre allocate enough
	// buffer of 32MB to cover majority of the cases.
	b.enc.Grow(32 << 20)
	return b
}

func (b *Batch) Add(tx KV, m *models.Model) error {
	id := tx.RecordID()
	domainHash := y.Hash(m.Domain)
	shard, ok := b.domains[domainHash]
	if !ok {
		shard = tx.Translate(v1.Field_domain, m.Domain)
		b.domains[domainHash] = shard
	}
	ts := uint64(time.UnixMilli(m.Timestamp).Truncate(time.Minute).UnixMilli())
	b.key.Time = ts
	b.key.Shard = uint32(shard)
	if m.Timestamp != 0 {
		b.get(v1.Field_timestamp).SetValue(id, m.Timestamp)
	}
	if m.Id != 0 {
		b.get(v1.Field_id).SetValue(id, int64(m.Id))
	}
	if m.Bounce != 0 {
		b.get(v1.Field_bounce).SetValue(id, int64(m.Bounce))
	}
	if m.Session {
		b.get(v1.Field_session).SetValue(id, 1)
	}
	if m.View {
		b.get(v1.Field_view).SetValue(id, 1)
	}
	if m.Duration != 0 {
		b.get(v1.Field_duration).SetValue(id, m.Duration)
	}
	if m.Duration != 0 {
		b.get(v1.Field_city).SetValue(id, int64(m.City))
	}
	b.set(tx, v1.Field_browser, id, m.Browser)
	b.set(tx, v1.Field_browser_version, id, m.BrowserVersion)
	b.set(tx, v1.Field_country, id, m.Country)
	b.set(tx, v1.Field_device, id, m.Device)
	// only store bitmap for domain. Use bsi existence field.
	b.get(v1.Field_domain).GetExistenceBitmap().Set(id)
	b.set(tx, v1.Field_domain, id, m.Domain)
	b.set(tx, v1.Field_entry_page, id, m.EntryPage)
	b.set(tx, v1.Field_event, id, m.Event)
	b.set(tx, v1.Field_exit_page, id, m.ExitPage)
	b.set(tx, v1.Field_host, id, m.Host)
	b.set(tx, v1.Field_os, id, m.Os)
	b.set(tx, v1.Field_os_version, id, m.OsVersion)
	b.set(tx, v1.Field_page, id, m.Page)
	b.set(tx, v1.Field_referrer, id, m.Referrer)
	b.set(tx, v1.Field_source, id, m.Source)
	b.set(tx, v1.Field_utm_campaign, id, m.UtmCampaign)
	b.set(tx, v1.Field_utm_content, id, m.UtmContent)
	b.set(tx, v1.Field_utm_medium, id, m.UtmMedium)
	b.set(tx, v1.Field_utm_source, id, m.UtmSource)
	b.set(tx, v1.Field_utm_term, id, m.UtmTerm)
	b.set(tx, v1.Field_subdivision1_code, id, m.Subdivision1Code)
	b.set(tx, v1.Field_subdivision2_code, id, m.Subdivision2Code)

	return nil
}

func (b *Batch) set(kv KV, field v1.Field, id uint64, value []byte) {
	if len(value) == 0 {
		return
	}
	b.get(field).SetValue(id, int64(kv.Translate(field, value)))
}

func (b *Batch) get(field v1.Field) *roaring.BSI {
	b.key.Field = field
	bs, ok := b.data[b.key]
	if !ok {
		bs = roaring.NewDefaultBSI()
		b.data[b.key] = bs
	}
	return bs
}

func (b *Batch) Save(db *badger.DB) (err error) {
	if len(b.data) == 0 {
		return
	}
	tx := db.NewTransaction(true)
	defer func() {
		if err != nil {
			tx.Discard()
		} else {
			err = tx.Commit()
		}
		clear(b.data)
		b.offsets = b.offsets[:0]
		b.enc.Reset()
	}()
	for k, v := range b.data {
		err = b.saveTs(tx, k, v)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				err = tx.Commit()
				if err != nil {
					return
				}
				tx = db.NewTransaction(true)
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

func (b *Batch) saveTs(tx *badger.Txn, key encoding.Key, value *roaring.BSI) error {
	ts := time.UnixMilli(int64(key.Time)).UTC()
	return errors.Join(
		b.saveKey(tx, encoding.Key{Field: key.Field}, value),                   // global
		b.saveKey(tx, encoding.Key{Field: key.Field, Shard: key.Shard}, value), // global by shard
		b.saveKey(tx, key, value),                                              // minute
		b.saveKey(tx, encoding.Key{Time: hour(ts), Shard: key.Shard, Field: key.Field}, value),
		b.saveKey(tx, encoding.Key{Time: day(ts), Shard: key.Shard, Field: key.Field}, value),
		b.saveKey(tx, encoding.Key{Time: week(ts), Shard: key.Shard, Field: key.Field}, value),
		b.saveKey(tx, encoding.Key{Time: month(ts), Shard: key.Shard, Field: key.Field}, value),
	)
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

func (b *Batch) saveKey(tx *badger.Txn, key encoding.Key, value *roaring.BSI) error {
	if key.Field == v1.Field_domain {
		return b.saveBitmap(
			tx,
			b.enc.Key(key),
			value,
		)
	}
	return b.save(
		tx,
		b.enc.Key(key),
		value,
	)
}

func (b *Batch) save(tx *badger.Txn, key []byte, value *roaring.BSI) error {
	it, err := tx.Get(key)
	var data []byte
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		b.offsets, data = value.Append(b.offsets[:0], b.enc.Allocate(value.GetSizeInBytes())[:0])
	} else {
		err = it.Value(func(val []byte) error {
			bs := roaring.NewBSIFromBuffer(val).Or(value)
			b.offsets, data = bs.Append(b.offsets[:0], b.enc.Allocate(bs.GetSizeInBytes())[:0])
			return err
		})
		if err != nil {
			return err
		}
	}
	return tx.Set(key, data)
}

func (b *Batch) saveBitmap(tx *badger.Txn, key []byte, value *roaring.BSI) error {
	it, err := tx.Get(key)
	var data []byte
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}

		data = value.GetExistenceBitmap().ToBuffer()
	} else {
		err = it.Value(func(val []byte) error {
			bs := roaring.Or(roaring.FromBuffer(val), value.GetExistenceBitmap())
			data = bs.ToBuffer()
			return err
		})
		if err != nil {
			return err
		}
	}
	return tx.Set(key, data)
}
