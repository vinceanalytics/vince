package oracle

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
	"go.etcd.io/bbolt"
)

type timeRange struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

type dbShard struct {
	mu     sync.RWMutex
	shards map[uint64]*db
	path   string
	db     *bbolt.DB
}

func newDBShard(path string) (*dbShard, error) {
	dirs, _ := os.ReadDir(path)
	o := &dbShard{shards: make(map[uint64]*db), path: path}
	if len(dirs) > 0 {
		for i := range dirs {
			if dirs[i].Name() == "TRANSLATE" {
				continue
			}
			shard, err := strconv.ParseUint(dirs[i].Name(), 10, 64)
			if err != nil {
				return nil, err
			}
			_, err = o.open(shard)
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
	for shard, s := range d.shards {
		start, end := s.min.Load(), s.max.Load()
		if max < start {
			continue
		}
		if min >= end {
			continue
		}
		rtx, err := s.db.Begin(false)
		if err != nil {
			return err
		}
		err = f(rtx, tx, shard)
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
	o, ok := d.shards[shard]
	d.mu.RUnlock()
	if ok {
		return o, nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	path := filepath.Join(d.path, formatShard(shard))
	x := rbf.NewDB(path, nil)
	err := x.Open()
	if err != nil {
		return nil, err
	}
	o = &db{
		db: x,
	}
	r, _ := os.ReadFile(filepath.Join(path, "time_range.json"))
	if len(r) > 0 {
		ts := timeRange{}
		json.Unmarshal(r, &ts)
		o.min.Store(ts.From)
		o.max.Store(ts.To)
	}
	d.shards[shard] = o
	return o, nil
}

type db struct {
	min, max atomic.Int64
	db       *rbf.DB
}

func (d *db) Close() error {
	data, _ := json.Marshal(timeRange{From: d.min.Load(), To: d.max.Load()})
	return errors.Join(
		os.WriteFile(filepath.Join(d.db.Path, "time_range.json"), data, 0600),
		d.db.Close(),
	)
}

func formatShard(shard uint64) string {
	return fmt.Sprintf("%06d", shard)
}
