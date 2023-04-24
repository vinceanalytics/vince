package timeseries

import (
	"context"
	"sync"
	"time"
)

// Maps user ID to *Buffer.
type Map struct {
	mu  sync.Mutex
	b   *bufMap
	ttl time.Duration
}

type bufMap struct {
	m       map[uint64]*Buffer
	deleted []uint64
}

func (b *bufMap) Release() {
	for k := range b.m {
		delete(b.m, k)
	}
	if len(b.deleted) > 0 {
		b.deleted = b.deleted[:0]
	}
}

func NewMap() *Map {
	return &Map{
		b: &bufMap{m: make(map[uint64]*Buffer)},
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
	m.mu.Lock()
	defer m.mu.Unlock()
	if b, ok := m.b.m[sid]; ok {
		return b
	}
	b := NewBuffer(uid, sid, m.ttl)
	m.b.m[sid] = b
	return b
}

// Removes the buffer associated with sid
func (m *Map) Delete(ctx context.Context, sid uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.b.deleted = append(m.b.deleted, sid)
}

func (m *Map) Save(ctx context.Context) {
	m.mu.Lock()
	x := m.b
	m.b = bufMapPool.Get().(*bufMap)
	m.mu.Unlock()
	defer x.Release()
	if len(m.b.deleted) == 0 {
		for _, v := range x.m {
			go Save(ctx, v)
		}
		return
	}
	h := make(map[uint64]struct{})
	for _, v := range m.b.deleted {
		h[v] = struct{}{}
	}
	for _, v := range x.m {
		if _, ok := h[v.SID()]; ok {
			// This site was deleted drop the buffer.
			v.Release()
		} else {
			go Save(ctx, v)
		}
	}
}

var bufMapPool = &sync.Pool{
	New: func() any {
		return &bufMap{m: make(map[uint64]*Buffer)}
	},
}
