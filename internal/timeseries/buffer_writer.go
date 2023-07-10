package timeseries

import (
	"context"
	"sync"
	"time"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/neo"
	"github.com/vinceanalytics/vince/pkg/entry"
)

type Buffer struct {
	mu     sync.Mutex
	neo    neo.ActiveBlock
	domain string
	ttl    time.Duration
}

func (b *Buffer) Init(domain string, ttl time.Duration) *Buffer {
	b.ttl = ttl
	b.domain = domain
	b.neo.Init(domain)
	return b
}

func (b *Buffer) Reset() *Buffer {
	b.neo.Reset()
	return b
}

func (b *Buffer) Release() {
	b.Reset()
	bigBufferPool.Put(b)
}

func NewBuffer(domain string, ttl time.Duration) *Buffer {
	return bigBufferPool.Get().(*Buffer).Init(domain, ttl)
}

func (b *Buffer) Register(ctx context.Context, e *entry.Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()
	e.Hit()
	x := caches.Session(ctx)
	cacheKey := e.ID
	if o, ok := x.Get(cacheKey); ok {
		s := o.(*entry.Entry)
		s.Update(e)
		defer e.Release()
	} else {
		x.SetWithTTL(e.ID, e, 1, b.ttl)
	}
	b.neo.WriteEntry(e)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return &Buffer{}
	},
}
