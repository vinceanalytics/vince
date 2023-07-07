package timeseries

import (
	"context"
	"sync"
	"time"

	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/pkg/entry"
)

type Buffer struct {
	mu     sync.Mutex
	domain string
	rows   []parquet.Row
	ttl    time.Duration
}

func (b *Buffer) endSession(e *entry.Entry) {
	e.IsBounce = e.EntryPage == e.ExitPage
	e.Value = 0
	b.rows = append(b.rows, e.Row())
}

func (b *Buffer) Init(domain string, ttl time.Duration) *Buffer {
	b.ttl = ttl
	b.domain = domain
	return b
}

func (b *Buffer) Reset() *Buffer {
	b.rows = b.rows[:0]
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
	b.rows = append(b.rows, e.Row())
	x := caches.Session(ctx)
	cacheKey := e.ID
	if o, ok := x.Get(cacheKey); ok {
		s := o.(*entry.Entry)
		s.Update(e)
		e.Release()
		return
	}
	x.SetWithTTL(e.ID, e, 1, b.ttl)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return &Buffer{}
	},
}
