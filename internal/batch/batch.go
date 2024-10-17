package batch

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/compute"
	"github.com/vinceanalytics/vince/internal/domains"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/util/hash"
)

const ShardWidth = 1 << 20

const textSize = models.Field_subdivision2_code - models.Field_browser

type Batch struct {
	data      map[encoding.Key]*roaring.BSI
	translate [textSize][][]byte
	cache     [textSize]map[uint32]struct{}
	key       encoding.Key
	enc       encoding.Encoding
	id        uint64
	shard     uint64
	txnCount  int
}

func NewBatch(db *badger.DB) *Batch {
	b := &Batch{
		data: make(map[encoding.Key]*roaring.BSI),
	}
	// a lot of small allocations happens during batching. We pre allocate enough
	// buffer of 32MB to cover majority of the cases.
	b.enc.Grow(32 << 20)
	db.View(func(txn *badger.Txn) error {
		key := b.enc.TranslateSeq(models.Field_unknown)
		it, err := txn.Get(key)
		if err != nil {
			return err
		}
		return it.Value(func(val []byte) error {
			b.id = binary.BigEndian.Uint64(val)
			return nil
		})
	})
	b.shard = b.id / ShardWidth
	return b
}

func (b *Batch) Add(m *models.Model) error {
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
	{
		// we special case handle domains. Normal hash32 translation requires
		// loading 32 bitmaps at worst case. We know domains cardinality is small
		// to speedup queries and reduce excess allocations. Domains translation
		// is done by domains.ID call.
		//
		// We never expose domain field to the end user so there is no need to care for
		// filter handles.
		//
		// By using site ID we ensure that for small views we will always only load
		// a single existence bitmap.A vince instance with 255 domains registerd
		/// will only need 9 bitmaps at worst case.
		did := domains.ID(string(m.Domain))
		b.get(models.Field_domain).SetValue(id, int64(did))
	}
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

func (b *Batch) set(field models.Field, id uint64, value []byte) {
	if len(value) == 0 {
		return
	}
	b.get(field).SetValue(id, b.tr(field, value))
}

func (b *Batch) get(field models.Field) *roaring.BSI {
	b.key.Field = field
	bs, ok := b.data[b.key]
	if !ok {
		bs = roaring.NewDefaultBSI()
		b.data[b.key] = bs
	}
	return bs
}

func (b *Batch) tr(field models.Field, value []byte) int64 {
	sum := hash.Sum32(value)
	idx := field.TextIndex()
	c := b.cache[idx]
	ok := c != nil
	if c == nil {
		c = make(map[uint32]struct{})
		b.cache[idx] = c
	}
	if ok {
		if _, has := c[sum]; has {
			return int64(sum)
		}
	}
	c[sum] = struct{}{}
	b.translate[idx] = append(b.translate[idx], value)
	return int64(sum)
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
		for f, v := range b.translate {
			clear(v)
			b.translate[f] = v[:0]
		}
		b.enc.Reset()
	}()

	set := func(key, value []byte) error {
		err := tx.Set(key, value)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				err = tx.Commit()
				if err != nil {
					return err
				}
				tx = db.NewTransaction(true)
				b.txnCount++
				err = tx.Set(key, value)
				if err != nil {
					return err
				}
				return nil
			}
			return err
		}
		return nil
	}

	// start by saving the current record id, even if some ops failed we will not
	//  mess up search
	seq := b.enc.Allocate(8)
	binary.BigEndian.PutUint64(seq, b.id)

	err = tx.Set(b.enc.TranslateSeq(models.Field_unknown), seq)
	if err != nil {
		return err
	}

	// save translations
	for idx, v := range b.translate {
		f := models.TextIndex(idx)
		for i := range v {
			sum := hash.Sum32(v[i])
			id := b.enc.TranslateID(f, sum)
			err := set(b.enc.TranslateKey(f, v[i]), nil)
			if err != nil {
				return fmt.Errorf("saving translation key %w", err)
			}
			err = set(id, v[i])
			if err != nil {
				return fmt.Errorf("saving translation value %w", err)
			}
		}
	}

	for k, v := range b.data {
		err = b.saveTs(tx, k, v)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				err = tx.Commit()
				if err != nil {
					return
				}
				tx = db.NewTransaction(true)
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

func (b *Batch) saveTs(tx *badger.Txn, key encoding.Key, value *roaring.BSI) error {
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

func (b *Batch) save(tx *badger.Txn, key []byte, value *roaring.Bitmap) error {
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
