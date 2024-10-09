package ro2

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dustin/go-humanize"
)

type DB struct {
	db *badger.DB
}

func newDB(path string) (*DB, error) {
	if path != "" {
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
	o := &DB{db: db}
	return o, nil
}

func (db *DB) Start(ctx context.Context) {
	go db.runVlogGC(ctx)
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Badger() *badger.DB { return db.db }

func (db *DB) Update(f func(tx *Tx) error) error {
	tx := newTx(nil)
	defer tx.Release()

	return db.db.Update(func(txn *badger.Txn) error {
		defer tx.Close()

		tx.tx = txn
		return f(tx)
	})
}

func (db *DB) View(f func(tx *Tx) error) error {
	return db.db.View(func(txn *badger.Txn) error {

		tx := newTx(txn)
		defer tx.Release()

		return f(tx)
	})
}

func (db *DB) runVlogGC(ctx context.Context) {
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
