package oracle

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
	"go.etcd.io/bbolt"
)

type dbShard struct {
	mu     sync.RWMutex
	shards []*db
	path   string
	db     *bbolt.DB
}

func newDBShard(path string) (*dbShard, error) {
	dirs, _ := os.ReadDir(path)
	o := &dbShard{shards: make([]*db, 0, 16), path: path}
	if len(dirs) > 0 {
		var shards []uint64
		for i := range dirs {
			if dirs[i].Name() == "TRANSLATE" {
				continue
			}
			shard, err := strconv.ParseUint(dirs[i].Name(), 10, 64)
			if err != nil {
				return nil, err
			}
			shards = append(shards, shard)
		}
		slices.Sort(shards)
		o.shards = make([]*db, 0, len(shards))
		for i := range shards {
			_, err := o.open(shards[i])
			if err != nil {
				o.Close()
				return nil, err
			}
		}
	}
	os.MkdirAll(path, 0755)
	tr := filepath.Join(path, "TRANSLATE")
	db, err := bbolt.Open(tr, 0600, nil)
	if err != nil {
		o.Close()
		return nil, err
	}
	o.db = db
	return o, nil
}

func (d *dbShard) Select(from, to int64, domain string,
	filter Filter,
	fn func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64, match *rows.Row) error) error {
	return d.viewDB(from, to, func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64) error {
		// 0: find domain
		domainField := newReadField(tx, []byte(domain))
		id, ok := domainField.get([]byte(domain))
		if !ok {
			return nil
		}
		f := rows.NewRow()
		var err error
		err = cursor.Tx(rTx, "domain", func(c *rbf.Cursor) error {
			f, err = cursor.Row(c, shard, id)
			return err
		})
		if err != nil {
			return err
		}
		if f.IsEmpty() {
			return nil
		}

		// 01: reduce domain rows with timestamp
		err = cursor.Tx(rTx, "timestamp", func(c *rbf.Cursor) error {
			f, err = compare(c, shard, orange, from, to, f)
			return err
		})
		if err != nil {
			return err
		}
		if f.IsEmpty() {
			return nil
		}
		match, err := filter.Apply(rTx, tx, shard, f)
		if err != nil {
			return err
		}
		if match.IsEmpty() {
			return nil
		}
		// f is the correct rows we want to read
		return fn(rTx, tx, shard, match)
	})
}

func (d *dbShard) viewDB(min, max int64, f func(rTx *rbf.Tx, tx *bbolt.Tx, shard uint64) error) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	tx, err := d.db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, s := range d.shards {
		if max < s.min.Load() {
			continue
		}
		if min >= s.max.Load() {
			continue
		}
		rtx, err := s.db.Begin(false)
		if err != nil {
			return err
		}
		err = f(rtx, tx, s.shard)
		rtx.Rollback()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dbShard) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	var errs []error
	for _, s := range d.shards {
		if e := s.Close(); e != nil {
			errs = append(errs, e)
		}
	}
	return errors.Join(errs...)
}

func (d *dbShard) open(shard uint64) (*db, error) {
	d.mu.RLock()
	i, ok := slices.BinarySearchFunc(d.shards, &db{shard: shard}, compareDB)
	if ok {
		o := d.shards[i]
		d.mu.RUnlock()
		return o, nil
	}
	d.mu.RUnlock()

	d.mu.Lock()
	defer d.mu.Unlock()
	path := filepath.Join(d.path, formatShard(shard))
	x := rbf.NewDB(path, nil)
	err := x.Open()
	if err != nil {
		return nil, err
	}
	o := &db{
		shard: shard,
		db:    x,
	}

	tx, err := x.Begin(false)
	if err != nil {
		o.Close()
		return nil, err
	}
	defer tx.Rollback()
	err = cursor.Tx(tx, "timestamp", func(c *rbf.Cursor) error {
		min, max, err := MinMax(c, shard)
		o.min.Store(min)
		o.max.Store(max)
		return err
	})
	if err != nil {
		o.Close()
		return nil, err
	}
	d.shards = slices.Insert(d.shards, i, o)
	return o, nil
}

type db struct {
	min, max atomic.Int64
	shard    uint64
	db       *rbf.DB
}

func compareDB(a, b *db) int {
	return cmp.Compare(a.shard, b.shard)
}

func (d *db) Close() error {
	return d.db.Close()
}

func formatShard(shard uint64) string {
	return fmt.Sprintf("%08d", shard)
}
