package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"runtime/trace"
	"sync"
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

// Iterate searches of all parquet files belonging to the user between start and
// end date. The emphasize here is start and end Must be dates.
//
// This requirements avoid further time/date conversions. Use timex.Date to get a
// date from time.Time.
//
// To stop iteration f must return ErrSkip.
func (b *Bob) Iterate(ctx context.Context, table TableID, user uint64, start, end time.Time, f func(io.ReaderAt, int64) error) error {
	_, task := trace.NewTask(ctx, "ts_iterate")
	defer task.End()

	return b.db.View(func(txn *badger.Txn) error {
		var id ID
		id.SetTable(table)
		id.SetUserID(user)
		id.SetDate(start)
		if start.Equal(end) {
			// This is an optimization. When we are iterating on data from the same
			// date. We know ID are prefixed with timestamp.
			//
			// We can just iterate on files for the user on this date.
			o := badger.DefaultIteratorOptions
			o.Prefix = id.PrefixWithDate()
			it := txn.NewIterator(o)
			defer it.Close()
			for ; it.Valid(); it.Next() {
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
		o.SinceTs = uint64(start.Unix())
		it := txn.NewIterator(o)

		// Iterate over all user's keys in this table.
		o.Prefix = id.Prefix()
		// Limit keys starting at beginning of the day on start date.
		o.SinceTs = uint64(start.Unix())

		// Limit keys up to the end of the day on end date.
		lastVersion := uint64(end.Unix())

		// we rely SinceTS value to seek to a better starting point for the iterator
		// SinceTS will be the time at the beginning of the day in which we want
		// to retrieve data from.
		for ; it.Valid(); it.Next() {
			x := it.Item()
			if x.Version() >= lastVersion {
				// we have reached the end of the iteration range.
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

// Merge  combines all the parquet files for today partitioned by user and site
// to a single file.
func (b *Bob) Merge(ctx context.Context) error {
	_, task := trace.NewTask(ctx, "ts_merge")
	defer task.End()

	start := time.Now()
	say := log.Get(ctx)
	say.Debug().Msg("starting merging daily parquet files")

	hash := make(map[uint64]uint64)

	defer func() {
		say.Debug().
			Int("users", len(hash)).
			Msgf("finished merging daily parquet files in %s", time.Since(start))
	}()

	// Try to find all sites which ingested events in a single day (Well, TODAY)
	b.db.View(func(txn *badger.Txn) error {
		var id ID
		id.SetTable(EVENTS)
		id.SetTime(time.Now())
		o := badger.DefaultIteratorOptions
		// we are only interested in keys only
		o.PrefetchValues = false
		o.Prefix = id[:userOffset]
		it := txn.NewIterator(o)
		defer it.Close()
		for it.Next(); it.Valid(); it.Next() {
			x := it.Item()
			if x.IsDeletedOrExpired() {
				continue
			}
			key := it.Item().Key()
			uid := binary.BigEndian.Uint64(key[userOffset:])
			sid := binary.BigEndian.Uint64(key[siteOffset:])
			hash[sid] = uid
		}
		return nil
	})

	merge := func(it *badger.Iterator, txn *badger.Txn, buf *Buffer) error {
		defer it.Close()
		for it.Next(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				f, err := parquet.OpenFile(bytes.NewReader(val), int64(len(val)))
				if err != nil {
					return err
				}
				g, err := parquet.MergeRowGroups(f.RowGroups())
				if err != nil {
					return err
				}
				parquet.CopyRows(buf.ew, g.Rows())
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
		}
		return nil
	}

	save := func(ctx context.Context, wg *sync.WaitGroup, uid, sid uint64) {
		defer wg.Done()
		err := b.db.Update(func(txn *badger.Txn) error {
			w := bigBufferPool.Get().(*Buffer).Init(uid, sid, 0)
			defer w.Release()

			w.id.SetDate(start)
			w.id.SetEntropy()

			o := badger.DefaultIteratorOptions
			o.Prefix = bytes.Clone(w.id[:entropyOffset])
			it := txn.NewIterator(o)
			err := merge(it, txn, w)
			if err != nil {
				return err
			}
			return w.save(txn, say)
		})
		if err != nil {
			log.Get(ctx).Err(err).
				Uint64("uid", uid).
				Uint64("sid", sid).
				Msg("failed merging events")
		}
	}
	if len(hash) > 0 {
		var wg sync.WaitGroup
		for sid, uid := range hash {
			wg.Add(1)
			go save(ctx, &wg, uid, sid)
		}
		wg.Wait()
		// try to reclaim space if possible.
		b.GC()
	}
	return nil

}
