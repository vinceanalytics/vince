package timeseries

import (
	"context"
	"encoding/binary"
	"errors"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/spec"
)

func Merge(ctx context.Context) {
	stats := storeForever(ctx)
	log.Get().Debug().
		Int("temporary", stats.keys.temporary).
		Int("permanent", stats.keys.permanent).
		Msgf("merged in %s", stats.elapsed)
}

type mergeStats struct {
	elapsed time.Duration
	keys    struct {
		temporary, permanent int
	}
}

// This is the heart of vince. Keys are read from the temporary store,
// transformed and stored in permanent storage.
func storeForever(ctx context.Context) (stats mergeStats) {
	start := core.Now(ctx)

	ts := uint64(start.Truncate(time.Hour).UnixMilli())

	readTs := uint64(start.UnixMilli())

	// Use this transaction to read from temporary storage
	tmpReadTxn := Temporary(ctx).NewTransactionAt(readTs, false)

	// Use this transaction to delete from temporary storage
	tmpDelTxn := Temporary(ctx).NewTransactionAt(readTs, true)

	// Transaction to store keys in permanent storage
	txn := Permanent(ctx).NewTransactionAt(ts, true)

	// To guarantee keys are not modified until committed we store them in this
	// buffer.
	buf := newSlice()

	var stamp [8]byte
	binary.BigEndian.PutUint64(stamp[:], ts)

	o := badger.DefaultIteratorOptions
	it := tmpReadTxn.NewIterator(o)

	// per iteration we have a single value with multiple keys
	var keys [][]byte

	for it.Rewind(); it.Valid(); it.Next() {
		x := it.Item()
		if x.IsDeletedOrExpired() {
			continue
		}
		stats.keys.temporary++

		key := x.Key()
		owk := buf.clone(key)
		err := doTxn(tmpDelTxn, readTs, tmpDelTxn.Delete(owk), func() error {
			tmpDelTxn = Temporary(ctx).NewTransactionAt(readTs, true)
			err := doTxn(txn, ts, badger.ErrTxnTooBig, func() error {
				txn = Permanent(ctx).NewTransactionAt(ts, true)
				return nil
			})
			buf.reset()
			return err
		})
		if err != nil {
			log.Get().Err(err).
				Str("key", formatID(key)).
				Msg("failed to commit delete or permanent store transaction")
			it.Close()
			txn.Discard()
			tmpDelTxn.Discard()
			tmpReadTxn.Discard()
			buf.release()
			return
		}
		keys = keys[:0]
		if owk[propOffset] == byte(spec.Base) {
			// Store global stats. Global stats are grouped into
			//  Per Site :
			//  Per User:
			//  Per Instance :
			// We also want to be able to chart or compute diffs between global stats so
			// we provide variations of per user and per instance stats with timestamp
			// appended.

			// we don't include BaseKey
			size := len(key) - len(spec.BaseKey)

			// plain stats
			g := buf.get(size)

			// plain stats by timestamp
			ts := buf.get(size)

			// set metric byte. We don't set property because this is property agnostic
			// and we want this to be sorted earlier to faster prefix iteration
			g[metricOffset] = key[metricOffset]

			ts[metricOffset] = key[metricOffset]
			copy(ts[len(ts)-8:], stamp[:]) //copy timestamp

			// per instance
			keys = append(keys, buf.clone(g), buf.clone(ts))

			// per user
			copy(g[:siteOffset], key[:siteOffset])
			copy(ts[:siteOffset], key[:siteOffset])
			keys = append(keys, buf.clone(g), buf.clone(ts))

			// per site
			copy(g[siteOffset:metricOffset], key[siteOffset:metricOffset])
			copy(ts[siteOffset:metricOffset], key[siteOffset:metricOffset])
			keys = append(keys, buf.clone(g), buf.clone(ts))
		} else {
			swk := buf.clone(key)
			copy(swk[len(swk)-8:], stamp[:])
			keys = append(keys, swk)
		}
		owv, err := x.ValueCopy(buf.get(8))
		if err != nil {
			log.Get().Err(err).
				Str("key", formatID(key)).
				Msg("failed to copy value from temporary store")
			it.Close()
			txn.Discard()
			tmpDelTxn.Discard()
			tmpReadTxn.Discard()
			buf.release()
			return
		}
		for _, swk := range keys {
			err := doTxn(txn, ts, saveKey(txn, buf, swk, owv), func() error {
				txn = Permanent(ctx).NewTransactionAt(ts, true)
				return doTxn(tmpDelTxn, ts, badger.ErrTxnTooBig, func() error {
					tmpDelTxn = Temporary(ctx).NewTransactionAt(readTs, true)
					buf.reset()
					return nil
				})
			})
			if err != nil {
				log.Get().Err(err).
					Str("key", formatID(swk)).
					Msg("failed to set key in permanent store")
				it.Close()
				txn.Discard()
				tmpDelTxn.Discard()
				tmpReadTxn.Discard()
				buf.release()
				return
			}
			stats.keys.permanent++
		}
	}
	it.Close()
	tmpReadTxn.Discard()
	err := tmpDelTxn.CommitAt(readTs, nil)
	if err != nil {
		log.Get().Err(err).Msg("failed to commit deletion of temporary keys transaction")
		tmpReadTxn.Discard()
		txn.Discard()
		buf.release()
		return
	}
	err = txn.CommitAt(ts, nil)
	if err != nil {
		log.Get().Err(err).Msg("failed to commit permanent storage transaction")
		tmpReadTxn.Discard()
		buf.release()
		return
	}
	return
}

func saveKey(txn *badger.Txn, s *slice, key, value []byte) error {
	item, err := txn.Get(key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return txn.Set(key, value)
		}
		return err
	}
	return item.Value(func(val []byte) error {
		return txn.Set(key, s.u64(binary.BigEndian.Uint64(val)+binary.BigEndian.Uint64(value)))
	})

}

func doTxn(txn *badger.Txn, stamp uint64, err error, on func() error) error {
	if err != nil {
		if errors.Is(err, badger.ErrTxnTooBig) {
			err = txn.CommitAt(stamp, nil)
			if err != nil {
				return err
			}
			return on()
		}
		return err
	}
	return nil
}
