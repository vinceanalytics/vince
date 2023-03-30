package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"runtime/trace"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/log"
	"github.com/golang/protobuf/proto"
	"golang.org/x/sync/errgroup"
)

var (
	ErrSkip = errors.New("skip iteration")
)

// Executed after stats have been merged to a single b
type MergeCallback func(ctx context.Context, b *Buffer, uid, sid uint64)

func Merge(ctx context.Context, since uint64, cb MergeCallback) (uint64, error) {
	_, task := trace.NewTask(ctx, "ts_merge")
	defer task.End()
	start := time.Now()
	say := log.Get(ctx)
	say.Debug().Msg("starting merging daily parquet files")
	db := GetBob(ctx)
	hash := make(map[uint64]uint64)

	defer func() {
		say.Debug().
			Int("users", len(hash)).
			Msgf("finished merging in %s", time.Since(start))
	}()

	// Try to find all sites which ingested events in a single day (Well, TODAY)
	db.View(func(txn *badger.Txn) error {
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = false
		// start with keys from our last processed key.
		o.SinceTs = since
		it := txn.NewIterator(o)
		defer it.Close()
		var last uint64
		for ; it.Valid(); it.Next() {
			x := it.Item()
			last = x.Version()
			if x.IsDeletedOrExpired() {
				continue
			}
			key := it.Item().Key()
			// only interested in site id and user id
			uid := binary.BigEndian.Uint64(key[userOffset:])
			sid := binary.BigEndian.Uint64(key[siteOffset:])
			hash[sid] = uid
		}
		// Track the last version we processed. Next merge will start from here.
		since = last
		return nil
	})

	if len(hash) == 0 {
		// early exit if there was no activity.
		return 0, nil
	}

	merge := func(it *badger.Iterator, txn *badger.Txn, buf *Buffer) error {
		defer it.Close()
		var data Entries
		de := getDecompressor()
		defer de.Release()
		for ; it.Valid(); it.Next() {
			item := it.Item()
			err := de.Read(item, func(val []byte) error {
				return proto.Unmarshal(val, &data)
			})
			if err != nil {
				return err
			}
			buf.entries = append(buf.entries, data.Events...)
			// delete the file, we are done merging it
			err = txn.Delete(item.Key())
			if err != nil {
				return err
			}
		}
		return nil
	}

	save := func(ctx context.Context, uid, sid uint64) func() error {
		return func() error {
			return db.Update(func(txn *badger.Txn) error {
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
				go cb(ctx, w.Sort(), uid, sid)
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
		db.RunValueLogGC(0.5)
	}
	return since, nil
}
