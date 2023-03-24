package timeseries

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/timex"
	"github.com/golang/protobuf/proto"
)

// Buffer buffers parquet file per user. Files are stored in badger db with keys
// of the form [TableID][UserID][Date][Random data], which is sortable. Basically
// we store daily stats and only care about the day [YY-MM-DD]. Individual files
// within the day are random.
//
// Since parquet data is sorted by timestamp, we can fast scan for relevant files/row
// groups by only relying on min/max of the timestamp column.
//
// This type is reusable.
type Buffer struct {
	expiresAt time.Time
	entries   []*Entry
	ttl       time.Duration
	mu        sync.Mutex
	id        ID
}

func (b *Buffer) Init(uid, sid uint64, ttl time.Duration) *Buffer {
	b.id.SetUserID(uid)
	b.id.SetSiteID(sid)
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
	for i := range b.id {
		b.id[i] = 0
	}
	b.expiresAt = time.Time{}
	b.entries = b.entries[:0]
	return b
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
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		e.SessionId = updated.SessionId
		b.entries = append(b.entries, updated, s, e)
		persist(ctx, updated)
		return
	}
	newSession := e.Session()
	e.SessionId = newSession.SessionId
	b.entries = append(b.entries, newSession, e)
	persist(ctx, newSession)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

func (b *Buffer) Save(ctx context.Context) error {
	say := log.Get(ctx)
	ts := Get(ctx)
	// data saved here is short lived
	b.id.Day(timex.Today())
	b.id.SetEntropy()
	ttl := b.ttl
	if ttl == 0 {
		// retain this for maximum of 3 hours merge window must always be less than this
		// to ensure data saved is processed.
		ttl = 3 * time.Hour
	}
	return ts.db.Update(func(txn *badger.Txn) error {
		x := Entries{
			Events: b.entries,
		}
		data, err := proto.Marshal(&x)
		if err != nil {
			return fmt.Errorf("failed to marshal events %v", err)
		}
		e := badger.NewEntry(b.id[:], data)
		if ttl != 0 {
			e.WithTTL(ttl)
		}
		err = txn.SetEntry(e)
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
