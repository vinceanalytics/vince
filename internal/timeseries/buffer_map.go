package timeseries

import (
	"context"
	"sync"
	"time"

	"github.com/vinceanalytics/vince/pkg/entry"
)

// Maps user ID to *Buffer.
type Map struct {
	m   map[string]*Buffer
	mu  sync.RWMutex
	ttl time.Duration
}

func NewMap(ttl time.Duration) *Map {
	return &Map{
		m:   make(map[string]*Buffer),
		ttl: ttl,
	}
}

type mapKey struct{}

func SetMap(ctx context.Context, m *Map) context.Context {
	return context.WithValue(ctx, mapKey{}, m)
}

func GetMap(ctx context.Context) *Map {
	return ctx.Value(mapKey{}).(*Map)
}

// Get returns a *Buffer belonging to a user with uid who owns site with id sid.
func (m *Map) Get(domain string) *Buffer {
	m.mu.RLock()
	b, ok := m.m[domain]
	if ok {
		m.mu.RUnlock()
		return b
	}
	m.mu.RUnlock()

	b = NewBuffer(domain, m.ttl)
	m.mu.Lock()
	m.m[domain] = b
	m.mu.Unlock()
	return b
}

// Removes the buffer associated with sid
func (m *Map) Delete(domain string) {
	m.mu.Lock()
	b, ok := m.m[domain]
	if ok {
		delete(m.m, domain)
		b.Release()
	}
	m.mu.Unlock()
}

func (m *Map) Close() error {
	m.mu.Lock()
	for k, v := range m.m {
		delete(m.m, k)
		v.Release()
	}
	m.mu.Unlock()
	return nil
}

func (m *Map) Save(ctx context.Context) {
	m.mu.Lock()
	for k, v := range m.m {
		delete(m.m, k)
		go Save(ctx, v)
	}
	m.mu.Unlock()
}

func Collect(ctx context.Context, e *entry.Entry) {
	GetMap(ctx).Get(e.Domain).endSession(e)
	e.Release()
}
