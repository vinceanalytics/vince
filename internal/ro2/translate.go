package ro2

import (
	"bytes"
	"encoding/binary"
	"errors"
	"log/slog"
	"math"
	"os"
	"slices"

	"github.com/dgraph-io/badger/v4"
)

const stringFieldSize = subdivision2_codeField - BrowserField + 1

type translator struct {
	tx     *Tx
	fields [stringFieldSize]uint64
	seen   [stringFieldSize]map[string]uint64
	key    []byte
	search []byte
	id     [2 + 8 + 8]byte
}

func (tr *translator) Reset(tx *Tx) {
	tr.key = slices.Grow(tr.key, 2+8+8)[:2+8+8]
	tr.search = slices.Grow(tr.key, 2+8)[:2+8]
	tr.key[0] = byte(TRANSLATE_KEY)
	tr.search[0] = byte(TRANSLATE_KEY)
	tr.id[0] = byte(TRANSLATE_ID)
	tr.tx = tx
}

func (tr *translator) init(tx *Tx) {
	it := tx.tx.NewIterator(badger.IteratorOptions{
		Reverse: true,
	})
	defer it.Close()
	tr.id[0] = byte(TRANSLATE_ID)
	for i := range tr.fields {
		f := uint64(i + BrowserField)
		binary.BigEndian.PutUint64(tr.id[2:], uint64(f))
		binary.BigEndian.PutUint64(tr.id[2+8:], math.MaxUint64)
		it.Seek(tr.id[:])
		ls := uint64(0)
		if it.Valid() {
			ls = binary.BigEndian.Uint64(it.Item().Key()[2:])
		}
		tr.fields[i] = ls
		tr.seen[i] = make(map[string]uint64)
	}
}

func (tr *translator) Tr(shard, field uint64, value string) (id uint64) {
	idx := field - BrowserField
	if v, ok := tr.seen[idx][value]; ok {
		return v
	}

	tr.key = tr.key[:10]
	binary.BigEndian.PutUint64(tr.key[2:], field)
	tr.key = append(tr.key, []byte(value)...)
	it, err := tr.tx.tx.Get(tr.key)
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
	tr.fields[idx]++
	id = tr.fields[idx]
	tr.seen[idx][value] = id

	binary.BigEndian.PutUint64(tr.id[2:], field)
	binary.BigEndian.PutUint64(tr.id[2+8:], id)

	// {field}-{id} => string(value)
	err = tr.tx.tx.Set(bytes.Clone(tr.id[:]), []byte(value))
	if err != nil {
		slog.Error("reading translation id", "err", err)
		os.Exit(1)
	}

	// {field}-{key} => uint64(id)
	err = tr.tx.tx.Set(bytes.Clone(tr.key), bytes.Clone(tr.id[2+8:]))
	if err != nil {
		slog.Error("saving translation key", "err", err)
		os.Exit(1)
	}

	// {field}-{shard}-{key} used for regex search
	tr.search = tr.search[:18]
	binary.BigEndian.PutUint64(tr.search[2:], field)
	binary.BigEndian.PutUint64(tr.search[2+8:], shard)
	tr.search = append(tr.search, []byte(value)...)

	err = tr.tx.tx.Set(bytes.Clone(tr.search), bytes.Clone(tr.id[2+8:]))
	if err != nil {
		slog.Error("saving  translation search key", "err", err)
		os.Exit(1)
	}
	return
}

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
