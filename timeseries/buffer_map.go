package timeseries

import (
	"context"
	"sync"
	"time"

	"github.com/gammazero/deque"
	"github.com/gernest/vince/log"
)

// Maps user ID to *Buffer.
type Map struct {
	ttl  time.Duration
	m    map[uint64]*Buffer
	mu   sync.Mutex
	ring *deque.Deque[*Buffer]
}

func NewMap(ttl time.Duration) *Map {
	return &Map{
		ttl:  ttl,
		m:    make(map[uint64]*Buffer),
		ring: deque.New[*Buffer](4098, 4098),
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
func (m *Map) Get(ctx context.Context, uid uint64) *Buffer {
	m.mu.Lock()
	defer m.mu.Unlock()
	var expired []*Buffer
	now := time.Now()
	for {
		if m.ring.Len() == 0 {
			break
		}
		e := m.ring.PopFront()
		if e.expired(now) {
			expired = append(expired, e)
			delete(m.m, e.id.UserID())
			continue
		}
		// We are sure nothing has expired yet from here on. Since we popped it from front
		// we return it up front. Old buffers must always be upfront.
		m.ring.PushFront(e)
		break
	}

	// release expired buffers in a separate goroutine.
	go m.release(ctx, expired...)

	if b, ok := m.m[uid]; ok {
		// found uid and it hasn't expired yet.
		return b
	}
	b := NewBuffer(uid, m.ttl)
	m.m[uid] = b
	m.ring.PushBack(b)
	return b
}

func (m *Map) release(ctx context.Context, b ...*Buffer) {
	x := log.Get(ctx)
	for _, v := range b {
		err := v.Save(ctx)
		if err != nil {
			x.Err(err).Msg("failed to save buffer")
		}
	}
}
