package dsl

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/rbf"
	"github.com/gernest/rbf/dsl/tx"
	"google.golang.org/protobuf/proto"
)

type Store[T proto.Message] struct {
	db     *rbf.DB
	ops    *Ops
	schema *Schema[T]

	shards roaring64.Bitmap

	mu sync.RWMutex
}

func New[T proto.Message](path string, bsi ...string) (*Store[T], error) {
	o, err := newOps(path)
	if err != nil {
		return nil, err
	}
	db := rbf.NewDB(filepath.Join(path, "rbf"), nil)
	err = db.Open()
	if err != nil {
		o.Close()
		return nil, err
	}

	schema, err := NewSchema[T](bsi...)
	if err != nil {
		o.Close()
		db.Close()
		return nil, err
	}

	s := &Store[T]{db: db, ops: o, schema: schema}

	// load shards
	txn, err := db.Begin(false)
	if err != nil {
		o.Close()
		db.Close()
		return nil, err
	}
	defer txn.Rollback()

	views := txn.FieldViews()
	prefix := tx.ViewKeyPrefix(ID)
	for i := range views {
		if strings.HasPrefix(views[i], prefix) {
			f := strings.TrimPrefix(views[i], prefix)
			f = strings.TrimSuffix(f, "<")
			shard, err := strconv.ParseUint(f, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parsing shards%w", err)
			}
			s.shards.Add(shard)
		}
	}
	return s, nil
}

func (s *Store[T]) Close() error {
	return errors.Join(s.db.Close(), s.ops.Close())
}

func (s *Store[T]) DB() *rbf.DB {
	return s.db
}

func (s *Store[T]) updateShards(shards []uint64) {
	s.mu.Lock()
	s.shards.AddMany(shards)
	s.mu.Unlock()
}

func (s *Store[T]) Shards() *roaring64.Bitmap {
	s.mu.RLock()
	a := s.shards.Clone()
	s.mu.RUnlock()
	return a
}

func (s *Store[T]) update(f func(tx *rbf.Tx) error) error {
	tx, err := s.db.Begin(true)
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
