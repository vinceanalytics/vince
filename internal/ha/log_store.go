package ha

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/must"
)

var _ raft.LogStore = (*DB)(nil)

var logPrefix = []byte{0x0}
var stablePrefix = []byte{0x1}

func (db *DB) FirstIndex() (v uint64, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: false,
			Reverse:        false,
		})
		defer it.Close()
		it.Seek(logPrefix)
		if it.ValidForPrefix(logPrefix) {
			v = binary.BigEndian.Uint64(it.Item().Key()[1:])
		}
		return nil
	})
	return
}

func (db *DB) LastIndex() (v uint64, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.IteratorOptions{
			PrefetchValues: false,
			Reverse:        true,
		})
		defer it.Close()

		it.Seek(append(logPrefix, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff))
		if it.ValidForPrefix(logPrefix) {
			v = binary.BigEndian.Uint64(it.Item().Key()[1:])
		}
		return nil
	})
	return
}

func (db *DB) GetLog(index uint64, log *raft.Log) error {
	return db.db.View(func(txn *badger.Txn) error {
		key := logKey(index)
		it, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return raft.ErrLogNotFound
			}
			return err
		}
		return it.Value(func(val []byte) error {
			dec := codec.NewDecoder(bytes.NewReader(val), &codec.BincHandle{})
			return dec.Decode(log)
		})
	})
}

func (db *DB) StoreLog(log *raft.Log) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(logKey(log.Index), serialize(log))
	})
}

func (db *DB) StoreLogs(logs []*raft.Log) error {
	return db.db.Update(func(txn *badger.Txn) error {
		err := make([]error, len(logs))
		for i, log := range logs {
			err[i] = txn.Set(
				logKey(log.Index),
				serialize(log),
			)
		}
		return errors.Join(err...)
	})
}

func (db *DB) DeleteRange(min, max uint64) error {
	// we manage the transaction manually in order to avoid ErrTxnTooBig errors
	txn := db.db.NewTransaction(true)
	it := txn.NewIterator(badger.IteratorOptions{
		PrefetchValues: false,
		Reverse:        false,
	})

	start := logKey(min)
	for it.Seek(start); it.Valid(); it.Next() {
		key := make([]byte, 9)
		it.Item().KeyCopy(key)
		// Handle out-of-range log index
		if binary.BigEndian.Uint64(key[1:]) > max {
			break
		}
		// Delete in-range log index
		if err := txn.Delete(key); err != nil {
			if err == badger.ErrTxnTooBig {
				it.Close()
				err = txn.Commit()
				if err != nil {
					return err
				}
				return db.DeleteRange(binary.BigEndian.Uint64(key[1:]), max)
			}
			return err
		}
	}
	it.Close()
	return txn.Commit()
}

func logKey(id uint64) []byte {
	var b [9]byte
	binary.BigEndian.PutUint64(b[1:], id)
	b[0] = logPrefix[0]
	return b[:]
}

func serialize(log *raft.Log) []byte {
	var b bytes.Buffer
	enc := codec.NewEncoder(&b, &codec.MsgpackHandle{})
	must.One(enc.Encode(log))("failed serializing log")
	return b.Bytes()
}
