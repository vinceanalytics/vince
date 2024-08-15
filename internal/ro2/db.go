package ro2

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dustin/go-humanize"
)

var (
	gc = flag.Duration("db.gc", time.Minute, "Interval for running GC checks on value log")
)

type DB struct {
	db *badger.DB
}

func newDB(path string) (*DB, error) {
	db, err := badger.Open(badger.DefaultOptions(path).WithLogger(nil))
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (db *DB) Start(ctx context.Context) {
	go db.runVlogGC(ctx)
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Update(f func(tx *Tx) error) error {
	tx := &Tx{}
	defer tx.release()
	return db.db.Update(func(txn *badger.Txn) error {
		tx.tx = txn
		return f(tx)
	})
}

func (db *DB) View(f func(tx *Tx) error) error {
	tx := &Tx{}
	defer tx.release()
	return db.db.View(func(txn *badger.Txn) error {
		tx.tx = txn
		return f(tx)
	})
}

func (db *DB) runVlogGC(ctx context.Context) {
	slog.Info("starting gc check loop")
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
