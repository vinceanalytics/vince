package ro2

import (
	"context"
	"encoding/binary"
	"flag"
	"log/slog"
	"math"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/dustin/go-humanize"
)

var (
	gc = flag.Duration("db.gc", time.Minute, "Interval for running GC checks on value log")
)

type DB struct {
	db     *badger.DB
	sysSeq atomic.Uint64
}

func newDB(path string) (*DB, error) {
	db, err := badger.Open(badger.
		DefaultOptions(path).
		WithCompression(options.ZSTD).
		WithLogger(nil))
	if err != nil {
		return nil, err
	}
	o := &DB{db: db}
	o.sysSeq.Store(o.latestID(dbSizeKey))
	return o, nil
}

func (db *DB) Start(ctx context.Context) {
	go db.runVlogGC(ctx)
	go db.runSystem(ctx)
}

func (o *DB) latestID(field uint64) (id uint64) {
	o.View(func(tx *Tx) error {
		key := tx.keys.Get()
		key.SetField(field).SetShard(math.MaxUint64)
		it := tx.tx.NewIterator(badger.IteratorOptions{
			Prefix:  key.KeyPrefix(),
			Reverse: true,
		})

		it.Rewind()
		if !it.Valid() {
			it.Close()
			return nil
		}
		shard := binary.BigEndian.Uint64(it.Item().Key()[shardOffset:])
		it.Close()

		// load existence bitmap for this shard
		exists := tx.Row(shard, timestampField, 0)
		if !exists.IsEmpty() {
			id = exists.Maximum()
		}
		return nil
	})
	return
}
func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Update(f func(tx *Tx) error) error {
	tx := txPool.Get().(*Tx)
	defer tx.Release()

	return db.db.Update(func(txn *badger.Txn) error {
		tx.tx = txn
		return f(tx)
	})
}

func (db *DB) View(f func(tx *Tx) error) error {
	tx := txPool.Get().(*Tx)
	defer tx.Release()

	return db.db.View(func(txn *badger.Txn) error {
		tx.tx = txn
		return f(tx)
	})
}

func (db *DB) runVlogGC(ctx context.Context) {
	slog.Info("starting gc check loop", "interval", gc.String())
	ticker := time.NewTicker(*gc)
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
