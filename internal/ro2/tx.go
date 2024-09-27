package ro2

import (
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/alicia"
)

//go:generate protoc  --go_out=. --go_opt=paths=source_relative pages.proto

type Tx struct {
	tx *badger.Txn
	it *badger.Iterator
	// we need to retain keys until the transaction is commited
	keys []*alicia.Key
}

var txPool = &sync.Pool{New: func() any {
	return &Tx{}
}}

func newTx(txn *badger.Txn) *Tx {
	tx := txPool.Get().(*Tx)
	tx.tx = txn
	return tx
}

func (tx *Tx) get() *alicia.Key {
	k := alicia.Get()
	tx.keys = append(tx.keys, k)
	return k
}

func (tx *Tx) Iter() *badger.Iterator {
	if tx.it == nil {
		tx.it = tx.tx.NewIterator(badger.IteratorOptions{})
	}
	return tx.it
}

func (tx *Tx) Close() {
	if tx.it != nil {
		tx.it.Close()
	}
	tx.it = nil
}

func (tx *Tx) Release() {
	tx.Close()
	tx.tx = nil
	for i := range tx.keys {
		tx.keys[i].Release()
	}
	clear(tx.keys)
	tx.keys = tx.keys[:0]
	txPool.Put(tx)
}
