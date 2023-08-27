package ha

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/hashicorp/raft"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ raft.LogStore = (*DB)(nil)

func (db *DB) FirstIndex() (v uint64, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		prefix := []byte(keys.RaftLog(-1) + "/")
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = false
		o.Prefix = prefix
		it := txn.NewIterator(o)
		defer it.Close()
		it.Rewind()
		if !it.Valid() {
			return raft.ErrLogNotFound
		}
		id := bytes.TrimPrefix(it.Item().Key(), prefix)
		v, err = strconv.ParseUint(string(id), 10, 64)
		return err
	})
	return
}

func (db *DB) LastIndex() (v uint64, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		prefix := []byte(keys.RaftLog(-1) + "/")
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = false
		o.Reverse = true
		o.Prefix = prefix
		it := txn.NewIterator(o)
		defer it.Close()
		it.Rewind()
		if !it.Valid() {
			return raft.ErrLogNotFound
		}
		id := bytes.TrimPrefix(it.Item().Key(), prefix)
		v, err = strconv.ParseUint(string(id), 10, 64)
		return err
	})
	return
}

func (db *DB) GetLog(index uint64, log *raft.Log) error {
	return db.db.View(func(txn *badger.Txn) error {
		key := []byte(keys.RaftLog(int64(index)))
		it, err := txn.Get(key)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return raft.ErrLogNotFound
			}
			return err
		}
		return it.Value(func(val []byte) error {
			var v v1.Raft_Log
			err := proto.Unmarshal(val, &v)
			if err != nil {
				return err
			}
			*log = raft.Log{
				Index:      v.Index,
				Term:       v.Term,
				Type:       raft.LogType(v.Type),
				Data:       v.Data,
				Extensions: v.Extensions,
				AppendedAt: v.AppendedAt.AsTime(),
			}
			return nil
		})
	})
}

func (db *DB) StoreLog(log *raft.Log) error {
	return db.db.Update(func(txn *badger.Txn) error {
		key := keys.RaftLog(int64(log.Index))
		return txn.Set([]byte(key), serialize(log))
	})
}

func (db *DB) StoreLogs(logs []*raft.Log) error {
	return db.db.Update(func(txn *badger.Txn) error {
		err := make([]error, len(logs))
		for i, log := range logs {
			err[i] = txn.Set(
				[]byte(keys.RaftLog(int64(log.Index))),
				serialize(log),
			)
		}
		return errors.Join(err...)
	})
}

func (db *DB) DeleteRange(min, max uint64) error {
	prefix := []byte(keys.RaftLog(-1))
	start := []byte(keys.RaftLog(int64(min)))
	end := []byte(keys.RaftLog(int64(max)))
	txn := db.db.NewTransaction(true)
	o := badger.DefaultIteratorOptions
	o.PrefetchValues = false
	o.Prefix = prefix
	it := txn.NewIterator(o)
	for it.Seek(start); it.Valid(); it.Next() {
		x := it.Item()
		key := x.Key()
		if bytes.Compare(end, key) == 1 {
			break
		}
		err := txn.Delete(key)
		if err != nil {
			if errors.Is(err, badger.ErrTxnTooBig) {
				it.Close()
				err = txn.Commit()
				if err != nil {
					return err
				}
				p := strings.Split(string(key), "/")
				id, _ := strconv.ParseUint(p[len(p)-1], 10, 64)
				return db.DeleteRange(id, max)
			}
			it.Close()
			txn.Discard()
			return err
		}
	}
	it.Close()
	return txn.Commit()
}

func serialize(log *raft.Log) []byte {
	return must.Must(proto.Marshal(&v1.Raft_Log{
		Index:      log.Index,
		Term:       log.Term,
		Type:       v1.Raft_Log_Type(log.Type),
		Data:       log.Data,
		Extensions: log.Extensions,
		AppendedAt: timestamppb.New(log.AppendedAt),
	}))("failed serializing raft log")
}
