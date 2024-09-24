package shards

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/vinceanalytics/vince/internal/rbf"
)

type DB struct {
	shards map[uint64]*rbf.DB
	mu     sync.RWMutex
	path   string
}

func New(path string) *DB {
	path = filepath.Join(path, "rbf")
	os.MkdirAll(path, 0755)
	return &DB{
		shards: make(map[uint64]*rbf.DB),
	}
}

func (db *DB) View(shard uint64, f func(tx *rbf.Tx) error) error {
	o := db.get(shard)
	tx, err := o.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	return f(tx)
}

func (db *DB) Update(shard uint64, f func(rtx *rbf.Tx) error) error {
	o := db.get(shard)
	tx, err := o.Begin(true)
	if err != nil {
		return err
	}
	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (db *DB) Close() error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	errs := make([]error, 0, len(db.shards))
	for _, o := range db.shards {
		errs = append(errs, o.Close())
	}
	return errors.Join(errs...)
}

func (db *DB) get(shard uint64) *rbf.DB {
	db.mu.RLock()
	o, ok := db.shards[shard]
	db.mu.RUnlock()
	if ok {
		return o
	}
	path := shardPath(db.path, shard)
	o = rbf.NewDB(path, nil)
	err := o.Open()
	if err != nil {
		slog.Error("opening database", "path", path, "err", err)
		os.Exit(1)
	}
	db.mu.Lock()
	db.shards[shard] = o
	db.mu.Unlock()
	return o
}

func shardPath(path string, shard uint64) string {
	return filepath.Join(
		path,
		fmt.Sprintf("%006d", shard),
	)
}
