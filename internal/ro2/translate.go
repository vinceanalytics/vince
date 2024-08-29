package ro2

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/dgraph-io/badger/v4"
)

func (tx *Tx) Tr(shard, field uint64, value string) (id uint64) {
	key := tx.get().TranslateKey(field, []byte(value))
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
	idKey := tx.get().TranslateID(field, id)
	err = tx.tx.Set(idKey, []byte(value))
	if err != nil {
		slog.Error("reading translation id", "err", err)
		os.Exit(1)
	}

	var b [8]byte
	binary.BigEndian.PutUint64(b[:], id)
	err = tx.tx.Set(key, b[:])
	if err != nil {
		slog.Error("saving translation key", "err", err)
		os.Exit(1)
	}
	return
}

func (tx *Tx) Next(field uint64) uint64 {
	key := tx.get().TranslateSeq(field)
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

func (tx *Tx) Find(field, id uint64) (o string) {
	key := tx.get().TranslateID(field, id)
	it, err := tx.tx.Get(key)
	if err != nil {
		fmt.Println("============")
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

func (tx *Tx) ID(field uint64, value string) (id uint64, ok bool) {
	key := tx.get().TranslateKey(field, []byte(value))
	it, err := tx.tx.Get(key)
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
	key := tx.get().TranslateKey(field, prefix)
	it := tx.tx.NewIterator(badger.IteratorOptions{
		Prefix: key,
	})
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		k := it.Item().Key()[2:]
		it.Item().Value(func(val []byte) error {
			f(k, binary.BigEndian.Uint64(val))
			return nil
		})
	}
}
