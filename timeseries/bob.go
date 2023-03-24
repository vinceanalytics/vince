package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"runtime/trace"
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
	cb    MergeCallback
	db    *badger.DB
	since uint64
}

func (b *Bob) GC() {
	b.db.RunValueLogGC(0.5)
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
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = false
		// start with keys from our last processed key.
		o.SinceTs = b.since
		it := txn.NewIterator(o)
		defer it.Close()
		var last uint64
		for it.Next(); it.Valid(); it.Next() {
			x := it.Item()
			last = x.Version()
			if x.IsDeletedOrExpired() {
				continue
			}
			key := it.Item().Key()
			uid := binary.BigEndian.Uint64(key[userOffset:])
			sid := binary.BigEndian.Uint64(key[siteOffset:])
			hash[sid] = uid
		}
		// Track the last version we processed. Next merge will start from here.
		b.since = last
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
				w.id.Day(start)
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
