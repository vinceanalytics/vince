package ha

import "github.com/dgraph-io/badger/v4"

type DB struct {
	db *badger.DB
}

func NewDB(db *badger.DB) *DB {
	return &DB{db: db}
}
