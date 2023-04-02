package timeseries

import (
	"bytes"
	"context"
	"errors"
	"runtime/trace"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/log"
	"github.com/golang/protobuf/proto"
)

var (
	ErrSkip = errors.New("skip iteration")
)

func Merge(ctx context.Context, since uint64, cb func(ctx context.Context, b *Buffer)) (uint64, error) {
	_, task := trace.NewTask(ctx, "ts_merge")
	defer task.End()
	say := log.Get(ctx)
	say.Debug().Msg("starting merging daily parquet files")
	db := GetBob(ctx)

	// Try to find all sites which ingested events in a single day (Well, TODAY)
	db.Update(func(txn *badger.Txn) error {
		o := badger.DefaultIteratorOptions
		o.AllVersions = true
		// we avoid sorting twice, rely on badger's sorting behavior.
		o.Reverse = true
		// start with keys from our last processed key.
		o.SinceTs = since
		it := txn.NewIterator(o)
		defer it.Close()
		var last uint64
		w := NewBuffer(0, 0, 0)
		first := true
		var data Entries

		for it.Rewind(); it.Valid(); it.Next() {
			x := it.Item()
			if x.IsDeletedOrExpired() {
				// skip expired keys
				continue
			}
			last = x.Version()
			if first {
				// first valid key to work with
				err := x.Value(func(val []byte) error {
					return proto.Unmarshal(val, &data)
				})
				if err != nil {
					return err
				}
				w.entries = append(w.entries, data.Events...)
				copy(w.id[:], x.Key())
				first = false
				continue
			}
			if !bytes.Equal(w.id[:], x.Key()) {
				// Finished iterating on this version of key send it for post
				// post processing
				send := w
				w = NewBuffer(0, 0, 0)
				go cb(ctx, send)
				copy(w.id[:], x.Key())
			}
			err := x.Value(func(val []byte) error {
				return proto.Unmarshal(val, &data)
			})
			if err != nil {
				return err
			}
			w.entries = append(w.entries, data.Events...)
		}

		if w.UID() == 0 {
			w.Release()
		} else {
			go cb(ctx, w)
		}
		// Track the last version we processed. Next merge will start from here.
		since = last
		return nil
	})
	return since, nil
}
