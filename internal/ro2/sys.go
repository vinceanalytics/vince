package ro2

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/vinceanalytics/vince/internal/ro"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

var (
	Requests atomic.Int64
)

const (
	// we don;t want keys for model to overlap with system keys
	dbSizeKey uint64 = iota + (1 << 10)
	requestsKey
	heapKey
)

var (
	sysInterval = flag.Duration("db.sys", 15*time.Minute, "Interval for collection system stats")
)

func (db *DB) runSystem(ctx context.Context) {
	slog.Info("starting db.sys collection loop", "interval", sysInterval.String())
	tick := time.NewTicker(*sysInterval)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-tick.C:
			err := db.collectSys(uint64(ts.UTC().UnixMilli()))
			if err != nil {
				slog.Error("collecting system stats", "err", err)
			}
		}
	}
}

func (db *DB) collectSys(ts uint64) error {
	shard := ts / ro.ShardWidth
	d, r, h := db.sysStats()
	database := roaring64.New()
	requests := roaring64.New()
	heap := roaring64.New()
	ro.BSI(database, ts, d)
	ro.BSI(requests, ts, r)
	ro.BSI(heap, ts, h)
	return db.Update(func(tx *Tx) error {
		return errors.Join(
			tx.Add(shard, dbSizeKey, nil, nil, database),
			tx.Add(shard, requestsKey, nil, nil, requests),
			tx.Add(shard, heapKey, nil, nil, heap),
		)
	})
}

func (db *DB) sysStats() (dbSize, requests, heap int64) {
	lsm, vlog := db.db.Size()
	dbSize = lsm + vlog
	requests = Requests.Load()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heap = int64(m.HeapAlloc)
	return
}
