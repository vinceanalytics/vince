package db

import (
	"cmp"
	"encoding/binary"
	"path/filepath"
	"slices"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/db"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/logger"
)

var (
	trKeys    = []byte("/tr/keys/")
	trIDs     = []byte("/tr/ids/")
	seqPrefix = []byte("/seq/")
)

type DB struct {
	db     *badger.DB
	shards *db.Shards
}

func New(path string) (*DB, error) {
	base := filepath.Join(path, "v1alpha1")
	shards := db.New(base)
	dbPth := filepath.Join(base, "db")
	db, err := badger.Open(badger.DefaultOptions(dbPth).WithLogger(nil))
	if err != nil {
		return nil, err
	}
	return &DB{db: db, shards: shards}, nil
}

func (db *DB) Close() error {
	var lastErr error
	if err := db.db.Close(); err != nil {
		lastErr = err
	}
	if err := db.shards.Close(); err != nil {
		lastErr = err
	}
	return lastErr
}

func (db *DB) UpdateShard(shard uint64, f func(tx *rbf.Tx) error) error {
	return db.shards.Update(shard, f)
}

func (db *DB) ViewShard(shard uint64, f func(tx *rbf.Tx) error) error {
	return db.shards.View(shard, f)
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

func (tx *Tx) find(prop string, key []byte) (uint64, bool) {
	hashKey := append(trKeys, []byte(prop)...)
	hashKey = append(hashKey, xxhash.New().Sum(key)...)
	if _, err := tx.Txn.Get(hashKey); err == nil {
		return xxhash.Sum64(key), true
	}
	return 0, false
}

func (tx *Tx) Cursor(name string, f func(c *rbf.Cursor) error) error {
	c, err := tx.Tx.Cursor(name)
	if err != nil {
		return err
	}
	defer c.Close()
	return f(c)
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
