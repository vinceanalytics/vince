package kv

import (
	"bytes"

	"github.com/cockroachdb/pebble"
)

type Pebble struct {
	db *pebble.DB
}

var _ KeyValue = (*Pebble)(nil)

func New(db *pebble.DB) *Pebble {
	return &Pebble{db: db}
}

func (db *Pebble) Set(key, value []byte) error {
	return db.db.Set(key, value, nil)
}

func (db *Pebble) Get(key []byte, value func([]byte) error) error {
	v, done, err := db.db.Get(key)
	if err != nil {
		return err
	}
	defer done.Close()
	return value(v)
}

func (db *Pebble) Prefix(key []byte, value func([]byte) error) error {
	it, err := db.db.NewIter(&pebble.IterOptions{
		LowerBound: key,
	})
	if err != nil {
		return err
	}
	defer it.Close()
	for it.First(); bytes.HasPrefix(it.Key(), key); it.Next() {
		err := value(bytes.TrimPrefix(it.Key(), key))
		if err != nil {
			return err
		}
	}
	return nil
}
