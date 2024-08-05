package sys

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/gernest/roaring/shardwidth"
	"github.com/vinceanalytics/vince/internal/assert"
	"github.com/vinceanalytics/vince/internal/rbf"
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
		case <-ts.C:
			Default.Apply(s.get())
		}
	}
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) get() (db *rbf.DB, shard uint64, f func(uint64) *rbf.DB) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db = s.db
	shard = s.shard
	f = s.open
	return
}

func (s *Store) open(shard uint64) *rbf.DB {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		err := s.db.Close()
		assert.Assert(err == nil, "closing system db", "err", err)
	}
	s.shard = shard
	s.db = rbf.NewDB(filepath.Join(s.path, fmt.Sprintf("%06d", shard)), nil)
	err := s.db.Open()
	assert.Assert(err == nil, "opening system db", "err", err)
	return s.db
}
