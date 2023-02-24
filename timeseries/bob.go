package timeseries

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/oklog/ulid/v2"
)

var (
	ErrSkip = errors.New("skip iteration")
)

// stores parquet files identified by ULID.
type Bob struct {
	db *badger.DB
}

func (b *Bob) GC() {
	b.db.RunValueLogGC(0.5)
}

type StoreRequest struct {
	Table string
	ID    ulid.ULID
	Data  []byte
	TTL   time.Duration
}

func (b *Bob) Store(r *StoreRequest) error {
	return b.db.Update(func(txn *badger.Txn) error {
		key := append([]byte(r.Table), r.ID[:]...)
		e := badger.NewEntry(key, r.Data)
		if r.TTL != 0 {
			e.WithTTL(r.TTL)
		}
		return txn.SetEntry(e)
	})
}

func CreateULID() ulid.ULID {
	return ulid.MustNew(
		ulid.Timestamp(toDate(time.Now())), ulid.DefaultEntropy(),
	)
}

// Reads parquet files in the range start .. end .Time comparison is done by date
func (b *Bob) Iterate(table string, start, end time.Time, f func(io.ReaderAt, int64) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		startDate := toDate(start)
		endDate := toDate(end)
		if startDate.Equal(endDate) {
			// This is an optimization. When we are iterating on data from the same
			// date. We know ULID are prefixed with timestamp. So the prefix for
			// the date will be append(table,id[:6]) where id is ulid with startDate
			// timestamp.
			var id ulid.ULID
			id.SetTime(ulid.Timestamp(startDate))
			o := badger.DefaultIteratorOptions
			prefix := append([]byte(table), id[:6]...)
			o.Prefix = prefix
			it := txn.NewIterator(o)
			defer it.Close()
			for it.Next(); it.Valid(); it.Next() {
				err := it.Item().Value(func(val []byte) error {
					return f(bytes.NewReader(val), int64(len(val)))
				})
				if err != nil {
					if errors.Is(err, ErrSkip) {
						return nil
					}
					return err
				}
			}
			return nil
		}
		o := badger.DefaultIteratorOptions
		o.Prefix = []byte(table)
		o.SinceTs = uint64(startDate.Unix())
		it := txn.NewIterator(o)

		var id, startTime, endTime ulid.ULID
		startTime.SetTime(ulid.Timestamp(startDate))
		endTime.SetTime(ulid.Timestamp(endDate))

		for ; it.Valid(); it.Next() {
			x := it.Item()
			copy(id[:], x.Key()[len(table):len(table)+6])
			// id,startTime and endTime all only contains timestamp part of the ulid
			// we check to see if startDate <= ulid <= endDate
			a := id.Compare(startTime)
			if a == -1 {
				// TOO early we can skip this iteration
				continue
			}
			if !within(a, id.Compare(endTime)) {
				break
			}
			err := x.Value(func(val []byte) error {
				return f(bytes.NewReader(val), int64(len(val)))
			})
			if err != nil {
				if errors.Is(err, ErrSkip) {
					return nil
				}
				return err
			}
		}
		return nil
	})
}

func within(a, b int) bool {
	return (a == 0 || a == 1) && (b == 0 || b == -1)
}
