package ro2

import (
	"encoding/binary"
	"errors"

	"github.com/dgraph-io/badger/v4"
)

func (tx *Tx) NextID() (id uint64, err error) {
	key := tx.get().Seq()
	it, err := tx.tx.Get(key)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			return 0, err
		}
	} else {
		err = it.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
		if err != nil {
			return
		}
	}
	id++
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	err = tx.tx.Set(key, b[:])
	return
}

func (tx *Tx) Seq() (id uint64) {
	key := tx.get().Seq()
	it, err := tx.tx.Get(key)
	if err == nil {
		it.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
	}
	return
}
