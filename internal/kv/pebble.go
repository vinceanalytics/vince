package kv

import "github.com/cockroachdb/pebble"

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
