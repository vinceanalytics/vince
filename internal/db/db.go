package db

import (
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl"
	"github.com/gernest/rbf/quantum"
	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/logger"
)

var (
	trKeys    = []byte("/tr/keys/")
	trIDs     = []byte("/tr/ids/")
	seqPrefix = []byte("/seq/")
)

type DB struct {
	db   *badger.DB
	path string

	seq *Seq

	shardsMu sync.RWMutex
	shards   map[uint64]*rbf.DB
}

func New(path string) (*DB, error) {
	base := filepath.Join("v1alpha1")
	dbPth := filepath.Join(base)
	db, err := badger.Open(badger.DefaultOptions(dbPth).WithLogger(nil))
	if err != nil {
		return nil, err
	}
	sq, err := NewSeq(db)
	if err != nil {
		db.Close()
		return nil, err
	}
	return &DB{db: db, seq: sq, shards: make(map[uint64]*rbf.DB)}, nil
}

func (db *DB) Close() error {
	var lastErr error
	if err := db.db.Close(); err != nil {
		lastErr = err
	}
	if err := db.seq.Release(); err != nil {
		lastErr = err
	}
	db.shardsMu.Lock()
	defer db.shardsMu.Unlock()
	for _, idx := range db.shards {
		err := idx.Close()
		if err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (db *DB) UpdateShard(shard uint64, f func(tx *rbf.Tx) error) error {
	return db.txShard(true, shard, f)
}

func (db *DB) ViewShard(shard uint64, f func(tx *rbf.Tx) error) error {
	return db.txShard(false, shard, f)
}

func (db *DB) txShard(update bool, shard uint64, f func(tx *rbf.Tx) error) error {
	idx, err := db.shard(shard)
	if err != nil {
		return err
	}
	tx, err := idx.Begin(update)
	if err != nil {
		return err
	}
	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !update {
		tx.Rollback()
		return nil
	}
	return tx.Commit()
}

func (db *DB) shard(shard uint64) (*rbf.DB, error) {
	db.shardsMu.RLock()
	idx, ok := db.shards[shard]
	db.shardsMu.RUnlock()
	if ok {
		return idx, nil
	}
	db.shardsMu.Lock()
	defer db.shardsMu.Unlock()

	idx = rbf.NewDB(filepath.Join(db.path, "index", strconv.FormatUint(shard, 10)), nil)
	err := idx.Open()
	if err != nil {
		return nil, err
	}
	db.shards[shard] = idx
	return idx, nil
}

type Seq struct {
	m sync.Map
}

func NewSeq(db *badger.DB) (*Seq, error) {
	seq := &Seq{}

	// all properties
	for i := v1.Property_event; i <= v1.Property_utm_term; i++ {
		q, err := db.GetSequence(append(seqPrefix, []byte(i.String())...), 4<<10)
		if err != nil {
			seq.Release()
			return nil, err
		}
		seq.m.Store(i.String(), q)
	}
	{
		// row ids
		q, err := db.GetSequence(append(seqPrefix, []byte("row_id")...), 4<<10)
		if err != nil {
			seq.Release()
			return nil, err
		}
		seq.m.Store("row_id", q)
	}
	return seq, nil
}

func (s *Seq) Release() (err error) {
	s.m.Range(func(key, value any) bool {
		x := value.(*badger.Sequence).Release()
		if x != nil {
			err = x
		}
		return true
	})
	return
}

func (s *Seq) NextRowID() (uint64, error) {
	q, err := s.get("row_id")
	if err != nil {
		return 0, err
	}
	return q.Next()
}

func (s *Seq) next(prop string) (uint64, error) {
	q, err := s.get(prop)
	if err != nil {
		return 0, err
	}
	return q.Next()
}

func (s *Seq) get(prop string) (*badger.Sequence, error) {
	q, ok := s.m.Load(prop)
	if !ok {
		return nil, fmt.Errorf("sequence %s not found", prop)
	}
	return q.(*badger.Sequence), nil
}

type Tx struct {
	Txn   *badger.Txn
	Shard uint64
	View  string
	Tx    *rbf.Tx
	DB    *DB
}

func (db *DB) View(f func(tx *Tx) error) error {
	return db.db.View(func(txn *badger.Txn) error {
		return f(&Tx{Txn: txn, DB: db})
	})
}

func (db *DB) Update(f func(tx *Tx) error) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return f(&Tx{Txn: txn, DB: db})
	})
}

func (db *DB) Append(events []*v1.Data) error {
	if len(events) == 0 {
		return nil
	}
	return db.Update(func(tx *Tx) error {
		return tx.add(events)
	})
}

func (tx *Tx) add(events []*v1.Data) error {
	if len(events) == 0 {
		return nil
	}
	rows, err := tx.DB.seq.get("rows_id")
	if err != nil {
		return err
	}
	m := map[string]*roaring.Bitmap{}

	get := func(k string) *roaring.Bitmap {
		b, ok := m[k]
		if !ok {
			b = roaring.NewBitmap()
			m[k] = b
		}
		return b
	}
	shards := map[string][]uint64{}

	err = Views(events, func(ts time.Time, events []*v1.Data) error {
		clear(m)
		var currShard uint64
		view := quantum.ViewByTimeUnit("", ts, 'D')
		for i, e := range events {
			id, err := rows.Next()
			if err != nil {
				return err
			}
			shard := id / shardwidth.ShardWidth
			if currShard != shard {
				currShard = shard
				if i != 0 {
					err := tx.save(view, shard, m)
					if err != nil {
						return err
					}
					clear(m)
				}
				shards[view] = append(shards[view], shard)
			}

			xid, err := tx.uid(uint64(e.Id))
			if err != nil {
				return err
			}
			dsl.AddBSI(get("ts"), id, uint64(e.Timestamp))
			dsl.AddBSI(get("uid"), id, xid)
			if e.Bounce == nil {
				dsl.AddBoolean(get("bounce"), id, false)
			} else {
				if *e.Bounce {
					dsl.AddBoolean(get("bounce"), id, true)
				}
			}
			if e.Duration != 0 {
				// convert to milliseconds
				ms := time.Duration(e.Duration * float64(time.Second)).Milliseconds()
				dsl.AddBSI(get("duration"), id, uint64(ms))
			}
			if e.Session {
				dsl.AddBoolean(get("session"), id, true)
			}
			if e.View {
				dsl.AddBoolean(get("view"), id, true)
			}
			tx.property(v1.Property_event, id, get, e.Event)
			tx.property(v1.Property_browser, id, get, e.Browser)
			tx.property(v1.Property_browser_version, id, get, e.BrowserVersion)
			tx.property(v1.Property_city, id, get, e.City)
			tx.property(v1.Property_country, id, get, e.Country)
			tx.property(v1.Property_device, id, get, e.Device)
			tx.property(v1.Property_domain, id, get, e.Domain)
			tx.property(v1.Property_entry_page, id, get, e.EntryPage)
			tx.property(v1.Property_exit_page, id, get, e.ExitPage)
			tx.property(v1.Property_host, id, get, e.Host)
			tx.property(v1.Property_os, id, get, e.Os)
			tx.property(v1.Property_os_version, id, get, e.OsVersion)
			tx.property(v1.Property_page, id, get, e.Page)
			tx.property(v1.Property_referrer, id, get, e.Referrer)
			tx.property(v1.Property_region, id, get, e.Region)
			tx.property(v1.Property_source, id, get, e.Source)
			tx.property(v1.Property_tenant_id, id, get, e.TenantId)
			tx.property(v1.Property_utm_campaign, id, get, e.UtmCampaign)
			tx.property(v1.Property_utm_content, id, get, e.UtmContent)
			tx.property(v1.Property_utm_medium, id, get, e.UtmMedium)
			tx.property(v1.Property_utm_source, id, get, e.UtmSource)
			tx.property(v1.Property_utm_term, id, get, e.UtmTerm)
		}
		return tx.save(view, currShard, m)
	})

	if err != nil {
		return err
	}
	// update view shard info
	for k, v := range shards {
		err := tx.saveViewInfo(k, v)
		if err != nil {
			return fmt.Errorf("saving view info for %s %w", k, err)
		}
	}
	return nil
}

func (tx *Tx) saveViewInfo(view string, shards []uint64) error {
	it, err := tx.Txn.Get([]byte(view))
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return err
		}
		return tx.Txn.Set([]byte(view), arrow.Uint64Traits.CastToBytes(shards))
	}
	it.Value(func(val []byte) error {
		a := arrow.Uint64Traits.CastFromBytes(val)
		shards = append(shards, a...)
		slices.Sort(shards)
		return nil
	})
	return tx.Txn.Set([]byte(view), arrow.Uint64Traits.CastToBytes(shards))
}

func (tx *Tx) save(view string, shard uint64, m map[string]*roaring.Bitmap) error {
	if len(m) == 0 {
		return nil
	}
	return tx.DB.UpdateShard(shard, func(tx *rbf.Tx) error {
		b := new(ViewFmt)
		for k, v := range m {
			_, err := tx.AddRoaring(b.Format(view, k), v)
			if err != nil {
				return fmt.Errorf("saving index for key %s %w", b.Format(view, k), err)
			}
		}
		return nil
	})
}

func (tx *Tx) property(prop v1.Property, id uint64, get func(string) *roaring.Bitmap, v string) {
	if v == "" {
		return
	}
	seq, err := tx.Upsert(prop, v)
	if err != nil {
		logger.Fail("translating key", "err", err)
	}
	dsl.AddBSI(get(prop.String()), id, seq)
}

func (tx *Tx) uid(v uint64) (uint64, error) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], v)
	return tx.upsert("uid", b[:])
}

func Views(events []*v1.Data, f func(ts time.Time, events []*v1.Data) error) error {
	if len(events) == 0 {
		return nil
	}
	slices.SortFunc(events, less)
	var i, j int
	ts := dateMS(events[0].Timestamp)
	valid := ts.UnixMilli()
	for ; j < len(events); j++ {
		if events[j].Timestamp < valid {
			continue
		}
		next := dateMS(events[j].Timestamp)
		switch ts.Compare(next) {
		case -1, 0:
		default:
			err := f(ts, events[i:j])
			if err != nil {
				return err
			}
			i = j
			ts = next
		}
	}
	return f(ts, events[i:])
}

func dateMS(ts int64) time.Time {
	return date(time.UnixMilli(ts))
}

func date(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func less(a, b *v1.Data) int {
	return cmp.Compare(a.Timestamp, b.Timestamp)
}

func (tx *Tx) Upsert(prop v1.Property, key string) (uint64, error) {
	return tx.upsert(prop.String(), []byte(key))
}

func (tx *Tx) find(prop string, key []byte) (uint64, bool) {
	hashKey := append(trKeys, []byte(prop)...)
	hashKey = append(hashKey, key...)
	if it, err := tx.Txn.Get(hashKey); err == nil {
		var id uint64
		err = it.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
		if err != nil {
			return 0, false
		}
		return id, true
	}
	return 0, false
}

func (tx *Tx) Tr(prop string, id uint64) (key string) {
	hashKey := append(trKeys, []byte(prop)...)
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	hashKey = append(hashKey, b[:]...)
	it, err := tx.Txn.Get(hashKey)
	if err != nil {
		logger.Fail("BUG: missing translation key", "prop", prop, "id", id)
	}
	it.Value(func(val []byte) error {
		key = string(val)
		return nil
	})
	return
}

func (tx *Tx) upsert(prop string, key []byte) (uint64, error) {
	hashKey := append(trKeys, []byte(prop)...)
	hashKey = append(hashKey, key...)
	if it, err := tx.Txn.Get(hashKey); err == nil {
		var id uint64
		err = it.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
		if err != nil {
			return 0, err
		}
		return id, nil
	}
	next, err := tx.DB.seq.next(prop)
	if err != nil {
		return 0, err
	}
	var id [8]byte
	binary.BigEndian.PutUint64(id[:], next)
	idKey := append(trIDs, []byte(prop)...)
	idKey = append(idKey, id[:]...)
	err = tx.Txn.Set(idKey, []byte(key))
	if err != nil {
		return 0, err
	}
	err = tx.Txn.Set(hashKey, id[:])
	if err != nil {
		return 0, err
	}
	return next, nil
}
