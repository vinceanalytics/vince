package ro2

import (
	"github.com/dgraph-io/badger/v4"
)

type DB struct {
	db *badger.DB
}

func New(path string) (*DB, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithLogger(nil))
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Update(f func(tx *Tx) error) error {
	tx := &Tx{}
	defer tx.release()
	return db.db.Update(func(txn *badger.Txn) error {
		tx.tx = txn
		return f(tx)
	})
}

func (db *DB) View(f func(tx *Tx) error) error {
	tx := &Tx{}
	defer tx.release()
	return db.db.View(func(txn *badger.Txn) error {
		tx.tx = txn
		return f(tx)
	})
}
