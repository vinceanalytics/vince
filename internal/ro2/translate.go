package ro2

import (
	"errors"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/models"
)

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

func (tx *Tx) Search(field models.Field, prefix []byte, f func([]byte)) {
	key := tx.enc.TranslateKey(field, nil)
	offset := len(key)
	key = append(key, prefix...)
	it := tx.Iter()
	for it.Seek(key); it.ValidForPrefix(key); it.Next() {
		k := it.Item().Key()[offset:]
		f(k)
	}
}
