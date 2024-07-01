package db

import (
	"encoding/binary"

	"github.com/cespare/xxhash/v2"
	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/logger"
)

type Translator struct {
	db     *badger.DB
	tx     *badger.Txn
	b      [8]byte
	h      xxhash.Digest
	update bool
}

func (t *Translator) Release() {
	if t.tx != nil {
		t.tx.Discard()
	}
}

func (t *Translator) Commit() error {
	if t.tx != nil {
		err := t.tx.Commit()
		t.tx = nil
		return err
	}
	return nil
}

func (t *Translator) Tr(prop string, key []byte) uint64 {
	if t.tx == nil {
		t.tx = t.db.NewTransaction(true)
	}
	hashKey := append(trKeys, []byte(prop)...)
	h, s := t.sum(key)
	hashKey = append(hashKey, s...)

	if _, err := t.tx.Get(hashKey); err == nil {
		return h
	}
	err := t.tx.Set(hashKey, key)
	if err != nil {
		logger.Fail("setting translation key", "err", err)
	}
	return h
}

func (t *Translator) sum(b []byte) (uint64, []byte) {
	t.h.Reset()
	t.h.Write(b)
	h := t.h.Sum64()
	binary.BigEndian.PutUint64(t.b[:], h)
	return h, t.b[:]
}
