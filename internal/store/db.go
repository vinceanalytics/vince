package store

import (
	"context"
	"encoding/binary"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/y"
	"github.com/dgraph-io/ristretto/z"
	"github.com/dgryski/go-farm"
	"github.com/dustin/go-humanize"
	"github.com/vinceanalytics/vince/internal/encoding"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/trie"
)

func newDB(path string) (*Store, error) {
	if path != "" {
		path = filepath.Join(path, "db")
		os.MkdirAll(path, 0755)
	}
	db, err := badger.Open(badger.
		DefaultOptions(path).
		WithInMemory(path == "").
		WithCompactL0OnClose(true).
		WithLogger(nil))
	if err != nil {
		return nil, err
	}
	store := &Store{
		db: db, tree: z.NewTree("translate"),
		trie: trie.NewTrie(),
	}
	for i := range store.mutex {
		store.mutex[i] = make(map[uint64]*roaring.Bitmap)
	}
	for i := range store.bsi {
		store.bsi[i] = make(map[uint64]*roaring.Bitmap)
	}
	for i := range store.ranges {
		f := models.Mutex(i)
		seq, err := db.GetSequence(encoding.TranslateSeq(f, make([]byte, 3)), maxTranslationKeyLeaseSize)
		y.Check(err)
		store.ranges[i] = seq
	}
	err = db.Update(func(txn *badger.Txn) error {
		{
			key := store.enc.TranslateSeq(models.Field_unknown)
			it, err := txn.Get(key)
			if err == nil {
				it.Value(func(val []byte) error {
					store.id = binary.BigEndian.Uint64(val)
					return nil
				})
			}

		}

		o := badger.DefaultIteratorOptions
		o.Prefix = keys.TranslateKeyPrefix
		o.PrefetchValues = false
		it := txn.NewIterator(o)
		slog.Info("loading translation")
		start := time.Now()
		count := 0
		defer func() {
			it.Close()
			slog.Info("loading translation complete", "keys", count, "elapsed", time.Since(start))
		}()
		for it.Seek(keys.TranslateKeyPrefix); it.Valid(); it.Next() {
			count++
			key := it.Item().Key()
			err := it.Item().Value(func(val []byte) error {
				id := binary.BigEndian.Uint64(val)
				store.tree.Set(farm.Fingerprint64(key), id)
				store.trie.Put(key, id)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if !errors.Is(err, badger.ErrKeyNotFound) {
		y.Check(err)
	}
	// a lot of small allocations happens during batching. We pre allocate enough
	// buffer of 32MB to cover majority of the cases.
	store.enc.Grow(32 << 20)
	store.shard = store.id / ShardWidth
	return store, nil
}

func (db *Store) Start(ctx context.Context) {
	go db.runVlogGC(ctx)
}

func (db *Store) Close() error {
	db.trie.Release()
	var errs []error
	errs = append(errs, db.Flush())
	for i := range db.ranges {
		errs = append(errs, db.ranges[i].Release())
	}
	errs = append(errs, db.db.Close())
	return errors.Join(errs...)
}

func (db *Store) Badger() *badger.DB { return db.db }

func (db *Store) Update(f func(tx *Tx) error) error {
	tx := db.newTx(nil)
	defer tx.Release()

	return db.db.Update(func(txn *badger.Txn) error {
		defer tx.Close()

		tx.tx = txn
		return f(tx)
	})
}

func (db *Store) View(f func(tx *Tx) error) error {
	return db.db.View(func(txn *badger.Txn) error {

		tx := db.newTx(txn)
		defer tx.Release()

		return f(tx)
	})
}

func (db *Store) runVlogGC(ctx context.Context) {
	slog.Info("starting gc check loop", "interval", time.Minute)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	abs := func(a, b int64) int64 {
		if a > b {
			return a - b
		}
		return b - a
	}

	var lastSz int64
	runGC := func() {
		for err := error(nil); err == nil; {
			// If a GC is successful, immediately run it again.
			err = db.db.RunValueLogGC(0.7)
		}
		_, sz := db.db.Size()
		if abs(lastSz, sz) > 512<<20 {
			slog.Info("Value log", "size", humanize.IBytes(uint64(sz)))
			lastSz = sz
		}
	}

	runGC()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runGC()
		}
	}
}
