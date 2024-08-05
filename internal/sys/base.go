package sys

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gernest/roaring/shardwidth"
	"github.com/vinceanalytics/vince/internal/assert"
	"github.com/vinceanalytics/vince/internal/rbf"
)

var Default = newSet()

type Counter struct {
	v atomic.Int64
}

func (c *Counter) Incr() {
	c.v.Add(1)
}

type Gauge struct {
	valueBits uint64
	f         func() float64
}

func (c *Gauge) Set(v float64) {
	assert.Assert(c.f == nil, "calling Set on a gauge with callback")
	n := math.Float64bits(v)
	atomic.StoreUint64(&c.valueBits, n)
}

func (c *Gauge) get() float64 {
	if c.f != nil {
		return c.f()
	}
	n := atomic.LoadUint64(&c.valueBits)
	return math.Float64frombits(n)
}

type set struct {
	mu      sync.RWMutex
	hist    map[string]*Histogram
	gauge   map[string]*Gauge
	counter map[string]*Counter
	cb      map[string]func()

	m *bitmaps
}

func newSet() *set {
	return &set{
		hist:    make(map[string]*Histogram),
		gauge:   make(map[string]*Gauge),
		counter: make(map[string]*Counter),
		cb:      map[string]func(){},
		m:       newBitmaps(),
	}
}

func (s *set) NewHistogram(name string) *Histogram {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.hist[name]
	assert.Assert(!ok, "histogram already exists", "name", name)
	h := new(Histogram)
	s.hist[name] = h
	return h
}

func (s *set) NewGauge(name string, f func() float64) *Gauge {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.gauge[name]
	assert.Assert(!ok, "gauge already exists", "name", name)
	h := &Gauge{f: f}
	s.gauge[name] = h
	return h
}

func (s *set) NewCounter(name string) *Counter {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.counter[name]
	assert.Assert(!ok, "counter already exists", "name", name)
	h := &Counter{}
	s.counter[name] = h
	return h
}

func (s *set) NewCallback(name string, f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.cb[name]
	assert.Assert(!ok, "callback already exists", "name", name)
	s.cb[name] = f
}

func (s *set) Apply(db *rbf.DB, shard uint64, change func(uint64) *rbf.DB) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, cb := range s.cb {
		cb()
	}
	ts := uint64(time.Now().UTC().UnixMilli())
	ns := ts / shardwidth.ShardWidth
	if ns != shard {
		db = change(ns)
	}
	b := s.m
	defer b.Reset()

	for n, v := range s.hist {
		b.Histogram(n, ts, v)
	}
	for n, v := range s.gauge {
		b.Gauge(n, ts, v)
	}
	for n, v := range s.counter {
		b.Counter(n, ts, v)
	}
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	for n, m := range b.m {
		_, err := tx.AddRoaring(n, m)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
