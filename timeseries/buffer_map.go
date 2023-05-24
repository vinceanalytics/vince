package timeseries

import (
	"context"
	"sync"
	"time"
)

// Maps user ID to *Buffer.
type Map struct {
	m   sync.Map
	ttl time.Duration
}

func NewMap() *Map {
	return &Map{}
}

type mapKey struct{}

func SetMap(ctx context.Context, m *Map) context.Context {
	return context.WithValue(ctx, mapKey{}, m)
}

func GetMap(ctx context.Context) *Map {
	return ctx.Value(mapKey{}).(*Map)
}

// Get returns a *Buffer belonging to a user with uid who owns site with id sid.
func (m *Map) Get(uid, sid uint64) *Buffer {
	if b, ok := m.m.Load(sid); ok {
		return b.(*Buffer)
	}
	b := NewBuffer(uid, sid, m.ttl)
	m.m.Store(sid, b)
	return b
}

// Removes the buffer associated with sid
func (m *Map) Delete(sid uint64) {
	if b, ok := m.m.LoadAndDelete(sid); ok {
		b.(*Buffer).Release()
	}
}

func (m *Map) Save(ctx context.Context) {
	m.m.Range(func(key, value any) bool {
		value.(*Buffer).Save(ctx)
		return true
	})
}
