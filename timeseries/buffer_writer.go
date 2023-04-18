package timeseries

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/log"
	"google.golang.org/protobuf/proto"
)

type Buffer struct {
	expiresAt time.Time
	entries   []*Entry
	ttl       time.Duration
	mu        sync.Mutex
	id        [16]byte
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	binary.BigEndian.PutUint64(b.id[:8], uid)
	binary.BigEndian.PutUint64(b.id[8:], sid)
	b.expiresAt = time.Now().Add(ttl)
	return b
}

func (b *Buffer) Sort() *Buffer {
	sort.Slice(b.entries, func(i, j int) bool {
		return b.entries[i].Timestamp < b.entries[j].Timestamp
	})
	return b
}

func (b *Buffer) Reset() *Buffer {
	b.expiresAt = time.Time{}
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

func (b *Buffer) expired(now time.Time) bool {
	return b.expiresAt.Before(now)
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

func (b *Buffer) Delete(ctx context.Context) {
	err := GetBob(ctx).Update(func(txn *badger.Txn) error {
		return txn.Delete(b.id[:])
	})
	if err != nil {
		log.Get(ctx).Err(err).
			Uint64("sid", b.SID()).
			Uint64("uid", b.UID()).
			Msg("failed to delete site collected events")
	}
	b.Release()
}

func (b *Buffer) Save(ctx context.Context) error {
	say := log.Get(ctx)
	ts := GetBob(ctx)
	// data saved here is short lived
	ttl := b.ttl
	if ttl == 0 {
		// retain this for maximum of 3 hours merge window must always be less than this
		// to ensure data saved is processed.
		ttl = 3 * time.Hour
	}
	return ts.Update(func(txn *badger.Txn) error {
		enc := getCompressor()
		defer enc.Release()

		x := Entries{
			Events: b.entries,
		}
		data, err := proto.Marshal(&x)
		if err != nil {
			return fmt.Errorf("failed to marshal events %v", err)
		}
		err = enc.Write(txn, b.id[:], data, ttl)
		if err != nil {
			say.Err(err).Msg("failed to save events to permanent storage")
			return err
		}
		say.Debug().Msgf("saved  %d entries ", len(b.entries))
		return nil
	})
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
