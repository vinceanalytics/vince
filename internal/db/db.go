package db

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	v2 "github.com/vinceanalytics/vince/gen/go/vince/v2"
	"go.etcd.io/bbolt"
)

type DB struct {
	db     *bbolt.DB
	idx    *rbf.DB
	batch  *Batch
	shards *roaring64.Bitmap
}

func New(path string) (*DB, error) {
	base := filepath.Join(path, "v1alpha1")
	db := rbf.NewDB(filepath.Join(base, "rbf"), nil)
	err := db.Open()
	if err != nil {
		return nil, err
	}
	tr, err := bbolt.Open(filepath.Join(base, "TRANSLATE"), 0600, nil)
	if err != nil {
		db.Close()
		return nil, err
	}
	shards := roaring64.New()
	if data, err := os.ReadFile(filepath.Join(db.Path, "SHARDS")); err == nil {
		err = shards.UnmarshalBinary(data)
		if err != nil {
			db.Close()
			tr.Close()
			return nil, err
		}
	}
	return &DB{db: tr, idx: db, batch: newBatch(), shards: shards}, nil
}

func (db *DB) Close() error {
	return errors.Join(db.Close(), db.idx.Close())
}

func (db *DB) Append(data *v2.Data) {
	db.batch.Append(data)
}

type view struct {
	tx  *rbf.Tx
	txn *bbolt.Tx
	m   map[string]*rbf.Cursor
}

func (c *view) Release() {
	for _, v := range c.m {
		v.Close()
	}
}

func (c *view) find(key, value string) (uint64, bool) {
	return find(c.txn, key, value)
}

func (c *view) get(name string) (*rbf.Cursor, error) {
	if v, ok := c.m[name]; ok {
		return v, nil
	}
	v, err := c.tx.Cursor(name)
	if err != nil {
		return nil, err
	}
	c.m[name] = v
	return v, nil
}
