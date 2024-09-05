package shards

import (
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
)

type Shards struct {
	shards []uint32
	minTs  []int64
	active atomic.Uint64

	mu sync.RWMutex
}

func New() *Shards {
	s := &Shards{}
	s.active.Store(^uint64(0))
	return s
}

func (s *Shards) Set(shards []uint32, ts []int64) {
	if len(shards) == 0 {
		s.active.Store(^uint64(0))
		return
	}
	s.mu.Lock()
	s.shards = shards
	s.minTs = ts
	s.mu.Unlock()
	s.active.Store(uint64(shards[len(shards)-1]))
}

func (s *Shards) Add(shard uint64, ts int64) (changed bool) {
	if s.active.Load() == shard {
		return false
	}
	s.active.Store(shard)
	s.mu.Lock()
	s.shards = append(s.shards, uint32(shard))
	s.minTs = append(s.minTs, ts)
	s.mu.Unlock()
	return true
}

func (s *Shards) Load() (shards []uint32, ts []int64) {
	s.mu.RLock()
	shards = slices.Clone(s.shards)
	ts = slices.Clone(s.minTs)
	s.mu.RUnlock()
	return
}

func (s *Shards) Select(start, end int64) (shards []uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.shards) == 0 {
		return []uint64{}
	}
	from, exactFrom := slices.BinarySearch(s.minTs, start)
	if !exactFrom {
		// adjust the index by going back one element. This ensures we are in the
		// right shard
		from--
	}
	if start == end {
		// we are looking only at a single shard. From has already been adjusted
		// no need to check exact match
		return []uint64{uint64(s.shards[from])}
	}
	to, _ := slices.BinarySearch(s.minTs, end)
	if from == to {
		return []uint64{uint64(s.shards[from])}
	}
	fmt.Println(from, to)
	shards = make([]uint64, 0, to-from)
	for i := from; i < to; i++ {
		shards = append(shards, uint64(s.shards[i]))
	}
	return
}
