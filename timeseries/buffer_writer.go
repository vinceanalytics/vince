package timeseries

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/gernest/vince/log"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go"
)

type Buffer struct {
	id       ID
	ttl      time.Duration
	eb       bytes.Buffer
	sb       bytes.Buffer
	ew       *parquet.SortingWriter[*Event]
	sw       *parquet.SortingWriter[*Session]
	sessions []*Session
	events   []*Event
}

func (b *Buffer) Init(user uint64, ttl time.Duration) *Buffer {
	b.id.SetUserID(user)
	b.ttl = ttl
	return b
}

func (b *Buffer) Reset() {
	for i := range b.id {
		b.id[i] = 0
	}
	b.ttl = 0
	b.eb.Reset()
	b.sb.Reset()
	b.events = b.events[:0]
	b.sessions = b.sessions[:0]
	bigBufferPool.Put(b)
}

func (b *Buffer) setup() *Buffer {
	b.ew = parquet.NewSortingWriter[*Event](&b.eb, SortRowCount, parquet.SortingWriterConfig(
		parquet.SortingColumns(
			parquet.Ascending("timestamp"),
		),
	))
	b.sw = parquet.NewSortingWriter[*Session](&b.eb, SortRowCount, parquet.SortingWriterConfig(
		parquet.SortingColumns(
			parquet.Ascending("timestamp"),
		),
	))
	return b
}

func NewBuffer(user uint64, ttl time.Duration) *Buffer {
	return bigBufferPool.Get().(*Buffer).Init(user, ttl)
}

func OnEvict(ctx context.Context) func(item *ristretto.Item) {
	return func(item *ristretto.Item) {
		b := item.Value.(*Buffer)
		b.Save(ctx)
	}
}

func (b *Buffer) Register(ctx context.Context, e *Event, prevUserId int64) uuid.UUID {
	var s *Session
	s = find(ctx, e, e.UserId)
	if s == nil {
		s = find(ctx, e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		b.sessions = append(b.sessions, updated, s)
		return persist(ctx, updated)
	}
	newSession := e.NewSession()
	b.sessions = append(b.sessions, newSession)
	b.events = append(b.events, e)
	return persist(ctx, newSession)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer).setup()
	},
}

func (b *Buffer) Save(ctx context.Context) {
	defer b.Reset()
	ts := Get(ctx)
	bob := &Bob{db: ts.db}
	b.id.SetTime(time.Now())
	b.id.Entropy()
	{
		// save events
		b.id.SetTable(EVENTS)
		err := b.ew.Close()
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to close parquet writer for events")
			return
		}
		err = bob.Store2(&b.id, b.eb.Bytes(), b.ttl)
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save events to permanent storage")
			return
		}
	}
	{
		// save sessions
		b.id.SetTable(SESSIONS)
		err := b.sw.Close()
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to close parquet writer for sessions")
			return
		}
		err = bob.Store2(&b.id, b.sb.Bytes(), b.ttl)
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save events to permanent storage")
			return
		}
	}
}
