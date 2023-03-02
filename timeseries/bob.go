package timeseries

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/vince/log"
	"github.com/segmentio/parquet-go"
)

var (
	ErrSkip = errors.New("skip iteration")
)

// Bob stores parquet files identified by ID.
type Bob struct {
	db *badger.DB
}

func (b *Bob) GC() {
	b.db.RunValueLogGC(0.5)
}

func storeTxn(id *ID, data []byte, ttl time.Duration, txn *badger.Txn) error {
	e := badger.NewEntry(id[:], data)
	if ttl != 0 {
		e.WithTTL(ttl)
	}
	return txn.SetEntry(e)
}

func (b *Bob) Iterate(table TableID, user uint64, start, end time.Time, f func(io.ReaderAt, int64) error) error {
	return b.db.View(func(txn *badger.Txn) error {
		startDate := toDate(start)
		endDate := toDate(end)
		if startDate.Equal(endDate) {
			// This is an optimization. When we are iterating on data from the same
			// date. We know ID are prefixed with timestamp.
			var id ID
			id.SetTable(table)
			id.SetUserID(user)
			id.SetTime(startDate)
			o := badger.DefaultIteratorOptions
			o.Prefix = id.PrefixWithDate()
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
		o.SinceTs = uint64(startDate.Unix())
		it := txn.NewIterator(o)

		var id, startTime, endTime ID

		startTime.SetTable(table)
		startTime.SetUserID(user)
		startTime.SetTime(startDate)
		endTime.SetTable(table)
		endTime.SetUserID(user)
		endTime.SetTime(endDate)
		o.Prefix = startTime.Prefix()

		// we rely SinceTS value to seek to a better starting point for the iterator
		// SinceTS will be the time at the beginning of the day in which we want
		// to retrieve data from.
		for ; it.Valid(); it.Next() {
			x := it.Item()
			copy(id[:], x.Key())
			// id,startTime and endTime all only contains timestamp part of the ulid
			// we check to see if startDate <= ulid <= endDate
			a := id.Compare(&startTime)
			if a == -1 {
				// Too early we can skip this iteration
				continue
			}
			if !within(a, id.Compare(&endTime)) {
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

// Merge  combines all the parquet files for today for a specific user
// to a single parquet file.
func (b *Bob) Merge(ctx context.Context) error {
	start := time.Now()
	say := log.Get(ctx)
	say.Debug().Msg("starting merging daily parquet files")

	var files int
	var id ID
	hash := make(map[uint64]struct{})
	id.SetTime(time.Now())

	defer func() {
		say.Debug().
			Int("users", len(hash)).
			Int("total_files", files).
			Msgf("finished merging daily parquet files in %s", time.Since(start))
	}()

	// Tries to find all users who ingested events in a single day (Well, TODAY)
	// We take advantage of SinceTs option for iterator for faster narrowing down
	// keys for the day.
	fetchUserID := func(t TableID, txn *badger.Txn) {
		// start with events
		id.SetTable(t)
		o := badger.DefaultIteratorOptions
		// we are only interested in keys
		o.PrefetchValues = false
		// perform table only iteration using seek to narrow down the time range
		now := toDate(time.Now()).Unix()
		o.Prefix = bytes.Clone(id[0:1])
		o.SinceTs = uint64(now)
		it := txn.NewIterator(o)
		defer it.Close()
		for it.Next(); it.Valid(); it.Next() {
			ts := id.Time().Unix()
			if now < ts {
				// version slightly order. Just skip them.
				continue
			}
			if now > ts {
				// not today
				break
			}
			copy(id[:], it.Item().Key())
			hash[id.UserID()] = struct{}{}
		}
	}
	b.db.View(func(txn *badger.Txn) error {
		fetchUserID(EVENTS, txn)
		// It is sufficient to just query the EVENTS table for active users in a
		// single Day. Basically same user stores one event and a possibility of
		// one or two sessions.
		//
		// This means it is safe to assume users found in a day in EVENTS table
		// are the same users with entries in SESSIONS table.
		return nil
	})

	// This is important!!! . Daily ingestion is rate limited per user, this guarantees
	// there won't be huge parquet data file  per user. We use in memory buffer instead of
	// a file for faster operations. The resulting file will be enough to safely
	// to store in badger.
	//
	// We are trading time for memory here. We use only this buffer for all merge
	// operations. Operations are performed sequentially, however it is safe to
	// call Merge concurrently. We offload database updates to badger transactions.
	w := NewBuffer(0, 0)
	defer w.Release()

	merge := func(t TableID, it *badger.Iterator, txn *badger.Txn) error {
		defer it.Close()
		for it.Next(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				f, err := parquet.OpenFile(bytes.NewReader(val), int64(len(val)))
				if err != nil {
					return err
				}
				switch t {
				case EVENTS:
					g, err := parquet.MergeRowGroups(f.RowGroups())
					if err != nil {
						return err
					}
					parquet.CopyRows(w.ew, g.Rows())
				case SESSIONS:
					g, err := parquet.MergeRowGroups(f.RowGroups())
					if err != nil {
						return err
					}
					parquet.CopyRows(w.sw, g.Rows())
				}
				return nil
			})
			if err != nil {
				return err
			}
			// delete the file, we are done merging it
			err = txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}

			// Track the number of files successfully merged
			files += 1
		}
		return nil
	}
	save := func(u uint64, txn *badger.Txn) error {
		w.Init(u, 0)
		defer w.Reset()
		for _, t := range []TableID{EVENTS, SESSIONS} {
			id.SetTable(t)
			o := badger.DefaultIteratorOptions
			o.Prefix = id.PrefixWithDate()
			it := txn.NewIterator(o)
			err := merge(t, it, txn)
			if err != nil {
				return err
			}
		}
		return w.save(txn, say)
	}
	if len(hash) > 0 {
		for u := range hash {
			// we are running per user transaction to avoid ErrTxnTooBig
			err := b.db.Update(func(txn *badger.Txn) error {
				return save(u, txn)
			})
			if err != nil {
				return err
			}
		}
		// try to reclaim space if possible.
		b.GC()
	}
	return nil

}
