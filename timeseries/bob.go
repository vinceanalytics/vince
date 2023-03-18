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
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/errgroup"
)

var (
	ErrSkip = errors.New("skip iteration")
)

// Bob stores parquet files identified by ID.
type Bob struct {
	cb MergeCallback
	db *badger.DB
}

func (b *Bob) GC() {
	b.db.RunValueLogGC(0.5)
}

// Iterate searches of all parquet files belonging to the user/site between start and
// end date. The emphasize here is start and end Must be dates.
//
// This requirements avoid further time/date conversions. Use timex.Date to get a
// date from time.Time.
//
// To stop iteration f must return ErrSkip.
func (b *Bob) Iterate(ctx context.Context, table TableID, uid, sid uint64, start, end time.Time, f func(io.ReaderAt, int64) error) {
	_, task := trace.NewTask(ctx, "ts_iterate")
	defer task.End()
	if start.Equal(end) {
		err := b.IterateDay(ctx, table, uid, sid, start, f)
		if err != nil {
			// TODO:(gernest) Include query in text format in this log context ?
			log.Get(ctx).Err(err).
				Uint64("uid", uid).
				Uint64("sid", sid).
				Msg("failed to iterate by day")
		}
		return
	}
	days := int(end.Sub(start).Hours() / 24)
	var wg sync.WaitGroup
	for i := 0; i < days; i += 1 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			date := start.AddDate(0, 0, n)
			if date.After(end) {
				return
			}
			err := b.IterateDay(ctx, table, uid, sid, date, f)
			if err != nil {
				// TODO:(gernest) Include query in text format in this log context ?
				log.Get(ctx).Err(err).
					Uint64("uid", uid).
					Uint64("sid", sid).
					Msg("failed to iterate by day")
			}
		}(i)
	}
	wg.Wait()
}

func (b *Bob) IterateDay(ctx context.Context, table TableID, uid, sid uint64, day time.Time, f func(io.ReaderAt, int64) error) error {
	_, task := trace.NewTask(ctx, "ts_iterate_day")
	defer task.End()
	return b.db.View(func(txn *badger.Txn) error {
		var id ID
		id.SetTable(table)
		id.SetUserID(uid)
		id.SetDate(day)
		id.SetSiteID(sid)
		o := badger.DefaultIteratorOptions
		o.Prefix = id[:entropyOffset]
		it := txn.NewIterator(o)
		defer it.Close()
		for ; it.Valid(); it.Next() {
			if ctx.Err() != nil {
				// support setting deadlines we don't want this to go on forever
				// for slower queries.
				return ctx.Err()
			}
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
	})
}

// Executed after stats have been merged to a single b
type MergeCallback func(ctx context.Context, b *Buffer, uid, sid uint64)

// Merge  combines all the parquet files for today partitioned by user and site
// to a single file.
//
// Merging is done in two steps. First we find all uid/sid keys crated during this
// merge window, then we process each unique uid/sid concurrently. By processing
// we mean merging together all parquet files into a single one.
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
		// we only merge last hour data.
		id.SetDayHour(time.Now().Add(-time.Hour))
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
		var data Entries
		for it.Next(); it.Valid(); it.Next() {
			err := it.Item().Value(func(val []byte) error {
				return proto.Unmarshal(val, &data)
			})
			if err != nil {
				return err
			}
			buf.entries = append(buf.entries, data.Events...)
			// delete the file, we are done merging it
			err = txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}
		}
		return nil
	}

	save := func(ctx context.Context, uid, sid uint64) func() error {
		return func() error {
			return b.db.Update(func(txn *badger.Txn) error {
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
				if b.cb != nil {
					b.cb(ctx, w.Sort(), uid, sid)
				}
				return nil
			})

		}
	}
	if len(hash) > 0 {
		g, ctx := errgroup.WithContext(ctx)
		for sid, uid := range hash {
			g.Go(save(ctx, uid, sid))
		}
		err := g.Wait()
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to merge ingested events")
		}
		b.GC()
	}
	return nil
}
