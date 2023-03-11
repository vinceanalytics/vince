package timeseries

import (
	"bytes"
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/docker/go-units"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/log"
	"github.com/rs/zerolog"
	"github.com/segmentio/parquet-go"
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
	mu        sync.Mutex
	id        ID
	expiresAt time.Time
	eb        bytes.Buffer
	sb        bytes.Buffer
	ew        *parquet.SortingWriter[*Entry]
}

func (b *Buffer) Init(user uint64, ttl time.Duration) *Buffer {
	b.id.SetUserID(user)
	b.expiresAt = time.Now().Add(ttl)
	return b
}

func (b *Buffer) Reset() *Buffer {
	for i := range b.id {
		b.id[i] = 0
	}
	b.expiresAt = time.Time{}
	b.eb.Reset()
	b.sb.Reset()
	return b
}

func (b *Buffer) Release() {
	b.Reset()
	bigBufferPool.Put(b)
}

func (b *Buffer) setup() *Buffer {
	eb := make([]parquet.BloomFilterColumn, len(eventsFilterFields))
	for i := range eventsFilterFields {
		eb[i] = parquet.SplitBlockFilter(10, eventsFilterFields[i])
	}
	b.ew = parquet.NewSortingWriter[*Entry](&b.eb, SortRowCount,
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		),
		parquet.BloomFilters(eb...),
	)
	return b
}

func NewBuffer(user uint64, ttl time.Duration) *Buffer {
	return bigBufferPool.Get().(*Buffer).Init(user, ttl)
}

func (b *Buffer) expired(now time.Time) bool {
	return b.expiresAt.Before(now)
}

func (b *Buffer) Register(ctx context.Context, e *Entry, prevUserId int64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var s *Entry
	s = find(ctx, e, e.UserId)
	if s == nil {
		s = find(ctx, e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = true
		s.Sign = false
		e.SessionId = updated.SessionId
		b.ew.Write([]*Entry{updated, s, e})
		persist(ctx, updated)
		return
	}
	newSession := e.Session()
	e.SessionId = newSession.SessionId
	b.ew.Write([]*Entry{newSession, e})
	persist(ctx, newSession)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer).setup()
	},
}

func (b *Buffer) Save(ctx context.Context) error {
	say := log.Get(ctx)
	ts := Get(ctx)
	b.id.SetTime(time.Now())
	b.id.Entropy()
	return ts.db.Update(func(txn *badger.Txn) error {
		return b.save(txn, say)
	})
}
func (b *Buffer) save(txn *badger.Txn, say *zerolog.Logger) error {
	b.id.SetTime(time.Now())
	b.id.Entropy()
	{
		// save events
		b.id.SetTable(EVENTS)
		err := b.ew.Close()
		if err != nil {
			say.Err(err).Msg("failed to close parquet writer for events")
			return err
		}
		err = storeTxn(&b.id, b.eb.Bytes(), 0, txn)
		if err != nil {
			say.Err(err).Msg("failed to save events to permanent storage")
			return err
		}
		say.Debug().Msgf("saved  %s events ", units.BytesSize(float64(b.eb.Len())))
	}
	return nil
}

func find(ctx context.Context, e *Entry, userId int64) *Entry {
	v, _ := caches.Session(ctx).Get(key(e.Domain, userId))
	if v != nil {
		return v.(*Entry)
	}
	return nil
}

func key(domain string, userId int64) string {
	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()
	b.WriteString(domain)
	b.WriteString(strconv.FormatInt(userId, 10))
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
