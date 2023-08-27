package ha

import "github.com/dgraph-io/badger/v4"

type DB struct {
	db *badger.DB
}
