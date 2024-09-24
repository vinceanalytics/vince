package dsl

import (
	"github.com/gernest/rbf/dsl/tr"
	"github.com/gernest/rbf/dsl/tx"
	"google.golang.org/protobuf/proto"
)

// Reader creates a new Reader for querying the store T. Make sure the reader is
// released after use.
func (s *Store[T]) Reader() (*Reader[T], error) {
	r, err := s.ops.read()
	if err != nil {
		return nil, err
	}
	return &Reader[T]{store: s, ops: r}, nil
}

type Reader[T proto.Message] struct {
	store *Store[T]
	ops   *readOps
}

func (r *Reader[T]) Release() error {
	return r.ops.Release()
}

func (r *Reader[T]) Tr() *tr.Read {
	return r.ops.tr
}

func (r *Reader[T]) View(f func(txn *tx.Tx) error) error {
	txn, err := r.store.db.Begin(false)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	it := r.store.Shards().Iterator()
	for it.HasNext() {
		err = f(tx.New(txn, it.Next(), r.ops.tr))
		if err != nil {
			return err
		}
	}
	return nil
}
