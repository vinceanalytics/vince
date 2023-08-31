package ha

import (
	"encoding/binary"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/keys"
)

var _ raft.StableStore = (*DB)(nil)

func (db *DB) Set(key, value []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (db *DB) Get(key []byte) (v []byte, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		key := keys.RaftStable(key)
		defer key.Release()
		it, err := txn.Get(key.Bytes())
		if err != nil {
			return err
		}
		v, err = it.ValueCopy(nil)
		return err
	})
	return
}

func (db *DB) SetUint64(key []byte, value uint64) error {
	var v [8]byte
	binary.BigEndian.PutUint64(v[:], value)
	return db.Set(key, v[:])
}

func (db *DB) GetUint64(key []byte) (uint64, error) {
	v, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}
