package sys

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/gernest/roaring"
	"github.com/gernest/roaring/shardwidth"
	"github.com/gernest/rows"
	"github.com/vinceanalytics/vince/internal/assert"
	"github.com/vinceanalytics/vince/internal/btx"
	"github.com/vinceanalytics/vince/internal/rbf"
	"github.com/vinceanalytics/vince/internal/rbf/cursor"
)

type Store struct {
	mu    sync.RWMutex
	shard uint64
	path  string
	db    *rbf.DB
}

func New(path string) (*Store, error) {
	os.MkdirAll(path, 0755)
	dirs, _ := os.ReadDir(path)
	s := &Store{
		path:  path,
		shard: uint64(time.Now().UTC().UnixMilli()) / shardwidth.ShardWidth,
	}
	if len(dirs) > 0 {
		shards := make([]uint64, len(dirs))
		for i := range dirs {
			shard, err := strconv.Atoi(dirs[i].Name())
			if err != nil {
				return nil, err
			}
			shards[i] = uint64(shard)
		}
		slices.Sort(shards)
		s.shard = shards[len(shards)-1]
	}
	s.open(s.shard)

	// Update counters
	tx, err := s.db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	exists, err := tx.RoaringBitmap("_id")
	if err != nil {
		return nil, err
	}

	if exists.Count() > 0 {
		r := rows.NewRowFromBitmap(exists)
		err := cursor.Tx(tx, "b0", func(c *rbf.Cursor) error {
			mx, _, err := btx.MaxUnsigned(c, r, s.shard, 64)
			b0.Store(mx)
			return err
		})
		if err != nil {
			return nil, err
		}
		err = cursor.Tx(tx, "b1", func(c *rbf.Cursor) error {
			mx, _, err := btx.MaxUnsigned(c, r, s.shard, 64)
			b1.Store(mx)
			return err
		})
		if err != nil {
			return nil, err
		}
		err = cursor.Tx(tx, "b2", func(c *rbf.Cursor) error {
			mx, _, err := btx.MaxUnsigned(c, r, s.shard, 64)
			b2.Store(mx)
			return err
		})
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Store) Start(ctx context.Context) {
	go s.start(ctx)
}

func (s *Store) start(ctx context.Context) {
	slog.Info("starting system metrics collection loop")
	ts := time.NewTicker(15 * time.Minute)
	defer ts.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ts.C:
			s.Apply(now)
		}
	}
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Apply(now time.Time) {
	ts := uint64(now.UnixMilli())
	shard := ts / shardwidth.ShardWidth

	s.mu.RLock()
	current := s.shard
	db := s.db
	s.mu.RUnlock()

	if shard != current {
		db = s.open(shard)
	}

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	heap := mem.HeapAlloc
	h0, h1, h2 := b0.Load(), b1.Load(), b2.Load()
	count := h0 + h1 + h2

	bitmap := roaring.NewBitmap()
	btx.BSI(bitmap, ts, int64(heap))

	tx, err := db.Begin(true)
	assert.Assert(err == nil, "starting write tx", "err", err)

	_, err = tx.Add("_id", btx.MutexPosition(ts, 0))

	assert.Assert(err == nil, "writing stats bitmap", "name", "err", err)

	_, err = tx.AddRoaring("heap", bitmap)
	assert.Assert(err == nil, "writing stats bitmap", "name", "err", err)

	{

		bitmap = roaring.NewBitmap()
		btx.BSI(bitmap, ts, int64(h0))
		_, err = tx.AddRoaring("b0", bitmap)
		assert.Assert(err == nil, "writing stats bitmap", "name", "err", err)

		bitmap = roaring.NewBitmap()
		btx.BSI(bitmap, ts, int64(h1))
		_, err = tx.AddRoaring("b1", bitmap)
		assert.Assert(err == nil, "writing stats bitmap", "name", "err", err)

		bitmap = roaring.NewBitmap()
		btx.BSI(bitmap, ts, int64(h2))
		_, err = tx.AddRoaring("b2", bitmap)
		assert.Assert(err == nil, "writing stats bitmap", "name", "err", err)

		bitmap = roaring.NewBitmap()
		btx.BSI(bitmap, ts, int64(count))
		_, err = tx.AddRoaring("count", bitmap)
		assert.Assert(err == nil, "writing stats bitmap", "name", "err", err)
	}

	err = tx.Commit()
	assert.Assert(err == nil, "saving stats bitmaps", "err", err)
}

func (s *Store) open(shard uint64) *rbf.DB {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		err := s.db.Close()
		assert.Assert(err == nil, "closing system db", "err", err)
	}
	s.shard = shard
	s.db = rbf.NewDB(filepath.Join(s.path, fmt.Sprintf("%08d", shard)), nil)
	err := s.db.Open()
	assert.Assert(err == nil, "opening system db", "err", err)
	return s.db
}

func (s *Store) Read() (*chartBSI, error) {
	s.mu.RLock()
	shard := s.shard
	db := s.db
	s.mu.RUnlock()
	tx, err := db.Begin(false)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	o := newChartBSI()
	ex, err := tx.RoaringBitmap("_id")
	if err != nil {
		return nil, err
	}
	f := rows.NewRowFromBitmap(ex)
	o.ts.AddMany(ex.Slice())
	err = cursor.Tx(tx, "heap", func(c *rbf.Cursor) error {
		return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
			o.ram.SetValue(column, value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	err = cursor.Tx(tx, "b0", func(c *rbf.Cursor) error {
		return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
			o.histograms[0].SetValue(column, value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	err = cursor.Tx(tx, "b1", func(c *rbf.Cursor) error {
		return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
			o.histograms[1].SetValue(column, value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	err = cursor.Tx(tx, "b2", func(c *rbf.Cursor) error {
		return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
			o.histograms[2].SetValue(column, value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	err = cursor.Tx(tx, "count", func(c *rbf.Cursor) error {
		return btx.ExtractBSI(c, shard, f, func(column uint64, value int64) error {
			o.requests.SetValue(column, value)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return o, nil
}
