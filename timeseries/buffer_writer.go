package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gernest/vince/caches"
)

type Buffer struct {
	entries []*Entry
	mu      sync.Mutex
	id      [16]byte
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	binary.BigEndian.PutUint64(b.id[:8], uid)
	binary.BigEndian.PutUint64(b.id[8:], sid)
	return b
}

func (b *Buffer) Clone() *Buffer {
	o := bufPool.Get().(*Buffer)
	copy(o.id[:], b.id[:])
	o.entries = append(o.entries, b.entries...)
	return o
}

// Clones b and save this to the data store in a separate goroutine. b is reset.
func (b *Buffer) Save(ctx context.Context) {
	b.mu.Lock()
	defer b.mu.Unlock()
	klone := b.Clone()
	b.entries = b.entries[:0]
	go Save(ctx, klone)
}

func (b *Buffer) Sort() *Buffer {
	sort.Slice(b.entries, func(i, j int) bool {
		return b.entries[i].Timestamp < b.entries[j].Timestamp
	})
	return b
}

func (b *Buffer) Reset() *Buffer {
	for _, e := range b.entries {
		e.Release()
	}
	b.entries = b.entries[:0]
	return b
}

func (b *Buffer) UID() uint64 {
	return binary.BigEndian.Uint64(b.id[:8])
}

func (b *Buffer) SID() uint64 {
	return binary.BigEndian.Uint64(b.id[8:])
}

func (b *Buffer) Release() {
	b.Reset()
	bigBufferPool.Put(b)
}

func NewBuffer(uid, sid uint64, ttl time.Duration) *Buffer {
	return bigBufferPool.Get().(*Buffer).Init(uid, sid, ttl)
}

func (b *Buffer) Register(ctx context.Context, e *Entry, prevUserId uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var s *Entry
	s = find(ctx, e, e.UserId)
	if s == nil {
		s = find(ctx, e, prevUserId)
	}
	if s != nil {
		// free e since we don't use it when doing updates
		defer e.Release()
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		b.entries = append(b.entries, updated, s)
		persist(ctx, updated)
		return
	}
	newSession := e.Session()
	b.entries = append(b.entries, newSession)
	persist(ctx, newSession)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

func find(ctx context.Context, e *Entry, userId uint64) *Entry {
	v, _ := caches.Session(ctx).Get(key(e.Domain, userId))
	if v != nil {
		return v.(*Entry)
	}
	return nil
}

func key(domain string, userId uint64) string {
	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()
	b.WriteString(domain)
	b.WriteString(strconv.FormatUint(userId, 10))
	return b.String()
}

var bufPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func persist(ctx context.Context, s *Entry) {
	caches.Session(ctx).SetWithTTL(key(s.Domain, s.UserId), s, 1, 30*time.Minute)
}
