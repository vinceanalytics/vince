package timeseries

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"
	"github.com/oklog/ulid/v2"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

var (
	IndexKeyPrefix = []byte("index/")
	ErrSkip        = errors.New("skip iteration")
)

type Bob struct {
	db *badger.DB
}

func (b *Bob) Create(name string) error {
	key := append(IndexKeyPrefix, []byte(name)...)
	return b.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		if err != nil {
			if !errors.Is(err, badger.ErrKeyNotFound) {
				return err
			}
		}
		b, err := proto.Marshal(&Store_Table{
			Name: name,
		})
		if err != nil {
			return err
		}
		return txn.Set(key, b)
	})
}

func (b *Bob) Store(table string, id ulid.ULID, data []byte, start, end time.Time) error {
	return b.db.Update(func(txn *badger.Txn) error {
		key := append(IndexKeyPrefix, []byte(table)...)
		t, err := b.table(txn, key)
		if err != nil {
			return err
		}
		begin := timestamppb.New(start)
		end := timestamppb.New(start)
		if t.Index == nil {
			t.Index = &Store_Index{}
		}
		t.Index.Entries = append(t.Index.Entries, &Store_Entry{
			Id: id[:],
			Range: &Store_Range{
				Start: begin,
				End:   end,
			},
		})
		if t.Index.Range == nil {
			t.Index.Range = &Store_Range{
				Start: begin,
				End:   end,
			}
		} else {
			t.Index.Range.End = end
		}
		err = txn.Set(id[:], data)
		if err != nil {
			return err
		}
		bs, err := proto.Marshal(t)
		if err != nil {
			return err
		}
		return txn.Set(key, bs)
	})
}

func (b *Bob) Iterate(table string, start, end time.Time, f func(io.ReaderAt, int64) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		key := append(IndexKeyPrefix, []byte(table)...)
		t, err := b.table(txn, key)
		if err != nil {
			return err
		}
		if t.Index == nil {
			return nil
		}
		if !t.Index.Range.Within(start, end) {
			return nil
		}
		for _, e := range t.Index.Entries {
			if e.Range.Within(start, end) {
				it, err := txn.Get(e.Id)
				if err != nil {
					return err
				}
				err = it.Value(func(val []byte) error {
					r := bytes.NewReader(val)
					return f(r, int64(len(val)))
				})
				if err != nil {
					if errors.Is(err, ErrSkip) {
						return nil
					}
					return err
				}
			}
		}
		return nil
	})
}

func (r *Store_Range) Within(start, end time.Time) bool {
	a := r.Start.AsTime()
	b := r.End.AsTime()
	if a.Before(start) && b.After(end) {
		return true
	}
	// earlier start date but ending within the boundary. This is same as
	// [r.Start, end]
	if start.Before(a) && b.After(end) {
		return true
	}
	return false
}

func (b *Bob) table(txn *badger.Txn, key []byte) (t *Store_Table, err error) {
	it, err := txn.Get(key)
	if err != nil {
		return nil, err
	}
	err = it.Value(func(val []byte) error {
		t = &Store_Table{}
		return proto.Unmarshal(val, t)
	})
	return
}
