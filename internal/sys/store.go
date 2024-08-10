package sys

import (
	"context"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
)

type Store struct {
	mu       sync.RWMutex
	heap     roaring64.BSI
	requests roaring64.BSI
	h0       roaring64.BSI
	h1       roaring64.BSI
	h2       roaring64.BSI
}

func New() *Store {
	return &Store{}
}

func (s *Store) Start(ctx context.Context, interval time.Duration) {
	go s.start(ctx, interval)
}

func (s *Store) start(ctx context.Context, interval time.Duration) {
	slog.Info("starting system metrics collection loop", "interval", interval)
	ts := time.NewTicker(interval)
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

func (s *Store) Apply(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ts := uint64(now.UnixMilli())
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	heap := mem.HeapAlloc
	h0, h1, h2 := b0.Load(), b1.Load(), b2.Load()
	count := h0 + h1 + h2
	s.heap.SetValue(ts, int64(heap))
	s.requests.SetValue(ts, int64(count))
	s.h0.SetValue(ts, int64(h0))
	s.h1.SetValue(ts, int64(h1))
	s.h2.SetValue(ts, int64(h2))
}
