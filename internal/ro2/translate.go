package ro2

import (
	"encoding/binary"
	"errors"
	"log/slog"
	"os"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/models"
)

func (tx *Tx) RecordID() uint64 {
	return tx.nextSeq(models.Field_unknown)
}

func (tx *Tx) Translate(field models.Field, value []byte) (id uint64) {
	key := tx.enc.TranslateKey(field, value)
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
	id = tx.nextSeq(field)
	idKey := tx.enc.TranslateID(field, id)
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

func (tx *Tx) nextSeq(field models.Field) uint64 {
	key := tx.enc.TranslateSeq(field)
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

func (tx *Tx) Find(field models.Field, id uint64) (o string) {
	key := tx.enc.TranslateID(field, id)
	it, err := tx.tx.Get(key)
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

func (tx *Tx) ids(field models.Field, value []string) []int64 {
	if len(value) == 0 {
		return []int64{}
	}
	if len(value) == 1 {
		id, ok := tx.ID(field, []byte(value[0]))
		if !ok {
			return []int64{}
		}
		return []int64{int64(id)}
	}
	o := make([]int64, 0, len(value))
	for i := range value {
		id, ok := tx.ID(field, []byte(value[i]))
		if !ok {
			continue
		}
		o = append(o, int64(id))
	}
	return o
}

func (tx *Tx) ID(field models.Field, value []byte) (id uint64, ok bool) {
	key := tx.enc.TranslateKey(field, value)
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

func (tx *Tx) Search(field models.Field, prefix []byte, f func([]byte, uint64)) {
	key := tx.enc.TranslateKey(field, nil)
	offset := len(key)
	key = append(key, prefix...)
	it := tx.Iter()

	for it.Seek(key); it.ValidForPrefix(key); it.Next() {
		k := it.Item().Key()[offset:]
		it.Item().Value(func(val []byte) error {
			f(k, binary.BigEndian.Uint64(val))
			return nil
		})
	}
}
