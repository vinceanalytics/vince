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

func (db *Pebble) Iter(low, upper []byte, value func([]byte) error) error {
	it, err := db.db.NewIter(&pebble.IterOptions{
		LowerBound: low,
		UpperBound: upper,
	})
	if err != nil {
		return err
	}
	defer it.Close()
	n := len(low)
	if len(upper) > 0 {
		n = len(upper)
	}
	trim := func(k []byte) []byte {
		if len(k) < n {
			return []byte{}
		}
		return k[n:]
	}
	for it.First(); bytes.HasPrefix(it.Key(), low); it.Next() {
		err := value(trim(it.Key()))
		if err != nil {
			return err
		}
	}
	return nil
}
