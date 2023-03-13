package timeseries

import (
	"context"
	"sync"
	"time"
)

// Maps user ID to *Buffer.
type Map struct {
	ttl time.Duration
	m   *sync.Map
}

func NewMap(ttl time.Duration) *Map {
	return &Map{
		ttl: ttl,
		m:   &sync.Map{},
	}
}

type mapKey struct{}

func SetMap(ctx context.Context, m *Map) context.Context {
	return context.WithValue(ctx, mapKey{}, m)
}

func GetMap(ctx context.Context) *Map {
	return ctx.Value(mapKey{}).(*Map)
}

// Get returns a *Buffer belonging to a user with uid. Expired buffers are released
// first before creating new one.
func (m *Map) Get(ctx context.Context, uid, sid uint64) *Buffer {
	if b, ok := m.m.Load(sid); ok {
		return b.(*Buffer)
	}
	b := NewBuffer(uid, sid, m.ttl)
	m.m.Store(sid, b)
	return b
}

func (m *Map) Save(ctx context.Context) {
	now := time.Now()
	var deleted []*Buffer
	m.m.Range(func(key, value any) bool {
		b := value.(*Buffer)
		if b.expired(now) {
			m.m.Delete(key.(uint64))
			deleted = append(deleted, b)
			return true
		}
		return true
	})
	go m.release(ctx, deleted...)
}

func (m *Map) release(ctx context.Context, x ...*Buffer) {
	for _, b := range x {
		b.Save(ctx)
		b.Release()
	}
}
