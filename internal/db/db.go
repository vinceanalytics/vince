package db

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/gernest/rbf"
	v2 "github.com/vinceanalytics/vince/gen/go/vince/v2"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

type DB struct {
	db     *bbolt.DB
	idx    *rbf.DB
	batch  *Batch
	ranges *v2.Shards
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
	ranges := &v2.Shards{}
	if data, err := os.ReadFile(filepath.Join(db.Path, "SHARDS")); err == nil {
		err := proto.Unmarshal(data, ranges)
		if err != nil {
			db.Close()
			tr.Close()
			return nil, err
		}
	}
	return &DB{db: tr, idx: db, batch: newBatch(), ranges: ranges}, nil
}

func (db *DB) Close() error {
	return errors.Join(db.db.Close(), db.idx.Close())
}

func (db *DB) Append(data *v2.Data) {
	db.batch.Append(data)
}

type view struct {
	tx    *rbf.Tx
	txn   *bbolt.Tx
	shard uint64
	m     map[string]*rbf.Cursor
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
	name = fmt.Sprintf("%s:%d", name, c.shard)
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

func (db *DB) updateShard(shard uint64, ts *minMax) {
	if len(db.ranges.Min) < int(shard) {
		db.ranges.Min = slices.Grow(db.ranges.Min, int(shard))[:shard]
		db.ranges.Max = slices.Grow(db.ranges.Max, int(shard))[:shard]
	}
	if db.ranges.Min[shard] == 0 {
		db.ranges.Min[shard] = ts.min
	} else {
		db.ranges.Min[shard] = min(db.ranges.Min[shard], ts.min)
	}
	db.ranges.Max[shard] = max(db.ranges.Max[shard], ts.max)
}
