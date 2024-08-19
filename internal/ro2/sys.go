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
	events atomic.Int64
)

const (
	// we don;t want keys for model to overlap with system keys
	dbSizeKey uint64 = iota + (1 << 10)
	requestsKey
	heapKey
	sysTs
)

var (
	sysInterval = flag.Duration("db.sys", 15*time.Minute, "Interval for collection system stats")
)

func Hit() {
	events.Add(1)
}

func (db *DB) Size() *roaring64.BSI {
	return db.sys(heapKey)
}

func (db *DB) Requests() *roaring64.BSI {
	return db.sys(requestsKey)
}

func (db *DB) Heap() *roaring64.BSI {
	return db.sys(heapKey)
}

func (db *DB) sys(field uint64) *roaring64.BSI {
	shard := db.sysSeq.Load() / ro.ShardWidth
	o := roaring64.NewDefaultBSI()
	err := db.View(func(tx *Tx) error {
		ts := readSysField(tx, shard, sysTs)
		t := readSysField(tx, shard, field)
		e := ts.GetExistenceBitmap().Iterator()
		for e.HasNext() {
			id := e.Next()
			stamp, _ := ts.GetValue(id)
			value, _ := t.GetValue(id)
			o.SetValue(uint64(stamp), value)
		}
		return nil
	})
	if err != nil {
		slog.Error("reading system metrics", "err", err)
	}
	return o
}

func (db *DB) runSystem(ctx context.Context) {
	slog.Info("starting db.sys collection loop", "interval", sysInterval.String())
	tick := time.NewTicker(*sysInterval)
	defer tick.Stop()
	err := db.collectSys(time.Now())
	if err != nil {
		slog.Error("collecting system stats", "err", err)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case ts := <-tick.C:
			err := db.collectSys(ts)
			if err != nil {
				slog.Error("collecting system stats", "err", err)
			}
		}
	}
}

func (db *DB) collectSys(now time.Time) error {
	id := db.sysSeq.Add(1)
	shard := id / ro.ShardWidth
	d, r, h := db.sysStats()
	database := roaring64.New()
	requests := roaring64.New()
	heap := roaring64.New()
	ts := roaring64.New()
	ro.BSI(database, id, d)
	ro.BSI(requests, id, r)
	ro.BSI(heap, id, h)
	ro.BSI(ts, id, now.UTC().UnixMilli())

	return db.Update(func(tx *Tx) error {
		return errors.Join(
			tx.Add(shard, dbSizeKey, nil, nil, database),
			tx.Add(shard, requestsKey, nil, nil, requests),
			tx.Add(shard, heapKey, nil, nil, heap),
			tx.Add(shard, sysTs, nil, nil, ts),
		)
	})
}

func readSysField(tx *Tx, shard, field uint64) *roaring64.BSI {
	o := roaring64.NewDefaultBSI()
	for i := uint64(0); i <= shard; i++ {
		tx.ExtractBSI(i, field, nil, func(row uint64, c int64) {
			o.SetValue(row, c)
		})
	}
	return o
}

func (db *DB) sysStats() (dbSize, requests, heap int64) {
	lsm, vlog := db.db.Size()
	dbSize = lsm + vlog
	requests = events.Load()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	heap = int64(m.HeapAlloc)
	return
}

func sysShard(ts time.Time) uint64 {
	yy, mm, dd := ts.Date()
	secs := time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).Unix()
	return uint64(secs) / ro.ShardWidth
}
