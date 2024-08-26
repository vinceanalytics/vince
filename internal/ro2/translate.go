package ro2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log/slog"
	"os"

	"github.com/dgraph-io/badger/v4"
)

func (tx *Tx) Tr(shard, field uint64, value string) (id uint64) {

	key := make([]byte, 2+8+len(value))
	key[0] = byte(TRANSLATE_KEY)
	binary.BigEndian.PutUint64(key[2:], field)
	copy(key[10:], []byte(value))
	it, err := tx.tx.Get(key)
	if err == nil {
		it.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
		return
	}

	if !errors.Is(err, badger.ErrKeyNotFound) {
		slog.Error("reading translation key", "err", err)
		os.Exit(1)
	}
	id = tx.Next(field)
	idKey := make([]byte, 2+8+8)
	idKey[0] = TRANSLATE_ID
	binary.BigEndian.PutUint64(idKey[2:], field)
	binary.BigEndian.PutUint64(idKey[2+8:], id)

	// {field}-{id} => string(value)
	err = tx.tx.Set(bytes.Clone(idKey), []byte(value))
	if err != nil {
		slog.Error("reading translation id", "err", err)
		os.Exit(1)
	}

	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	// {field}-{string} => uint64(id)
	err = tx.tx.Set(key, b[:])
	if err != nil {
		slog.Error("saving translation key", "err", err)
		os.Exit(1)
	}

	return
}

func (tx *Tx) Next(field uint64) uint64 {
	key := make([]byte, 2+8)
	key[0] = TRANSLATE_SEQ
	binary.BigEndian.PutUint64(key[2:], field)
	var id uint64
	it, err := tx.tx.Get(key)
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading translation sequence  key", "err", err)
			os.Exit(1)
		}
	} else {
		it.Value(func(val []byte) error {
			id = binary.BigEndian.Uint64(val)
			return nil
		})
	}
	id++
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	err = tx.tx.Set(key, b[:])
	if err != nil {
		slog.Error("saving translation sequence  key", "err", err)
		os.Exit(1)
	}
	return id
}

func (tx *Tx) Find(field, key uint64) (o string) {
	var buf [2 + 8 + 8]byte
	buf[0] = byte(TRANSLATE_ID)
	binary.BigEndian.PutUint64(buf[2:], field)
	binary.BigEndian.PutUint64(buf[2+8:], key)
	it, err := tx.tx.Get(buf[:])
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading translation key", "key", err, "err", err)
		}
		return
	}
	it.Value(func(val []byte) error {
		o = string(val)
		return nil
	})
	return
}

func (tx *Tx) ID(field uint64, key string) (id uint64, ok bool) {
	buf := make([]byte, 10+len(key))
	buf[0] = byte(TRANSLATE_KEY)
	binary.BigEndian.PutUint64(buf[2:], field)
	copy(buf[10:], []byte(key))
	it, err := tx.tx.Get(buf[:])
	if err != nil {
		if !errors.Is(err, badger.ErrKeyNotFound) {
			slog.Error("reading translation key", "key", err, "err", err)
		}
		return
	}
	it.Value(func(val []byte) error {
		id = binary.BigEndian.Uint64(val)
		ok = true
		return nil
	})
	return
}

func (tx *Tx) Search(field uint64, prefix []byte, f func([]byte, uint64)) {
	key := make([]byte, 2+8+len(prefix))
	key[0] = byte(TRANSLATE_KEY)
	binary.BigEndian.PutUint64(key[2:], field)
	copy(key[10:], prefix)

	it := tx.tx.NewIterator(badger.IteratorOptions{
		Prefix: key,
	})
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		k := it.Item().Key()[10:]
		it.Item().Value(func(val []byte) error {
			f(k, binary.BigEndian.Uint64(val))
			return nil
		})
	}
}
