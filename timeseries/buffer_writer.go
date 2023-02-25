package timeseries

import (
	"bytes"
	"context"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/docker/go-units"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/log"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go"
)

type Buffer struct {
	id       ID
	start    time.Time
	ttl      time.Duration
	eb       bytes.Buffer
	sb       bytes.Buffer
	ew       *parquet.SortingWriter[*Event]
	sw       *parquet.SortingWriter[*Session]
	sessions [2]*Session
	events   [1]*Event
}

func (b *Buffer) Init(user uint64, ttl time.Duration) *Buffer {
	b.id.SetUserID(user)
	b.ttl = ttl
	b.start = time.Now()
	return b
}

func (b *Buffer) Reset() {
	for i := range b.id {
		b.id[i] = 0
	}
	b.ttl = 0
	b.eb.Reset()
	b.sb.Reset()
	b.events[0] = nil
	b.sessions[0] = nil
	b.sessions[1] = nil
}

func (b *Buffer) Release() {
	b.Reset()
	bigBufferPool.Put(b)
}

func (b *Buffer) setup() *Buffer {
	b.ew = parquet.NewSortingWriter[*Event](&b.eb, SortRowCount, parquet.SortingWriterConfig(
		parquet.SortingColumns(
			parquet.Ascending("timestamp"),
		),
	))
	b.sw = parquet.NewSortingWriter[*Session](&b.sb, SortRowCount, parquet.SortingWriterConfig(
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

func (b *Buffer) Expired() bool {
	return b.start.Add(b.ttl).Before(time.Now())
}

func OnReject(item *ristretto.Item) {
	b := item.Value.(*Buffer)
	b.Release()
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
		b.sessions[0] = updated
		b.sessions[1] = s
		b.sw.Write(b.sessions[:])
		return persist(ctx, updated)
	}
	newSession := e.NewSession()
	b.sessions[0] = newSession
	b.sw.Write(b.sessions[:1])
	b.events[0] = e
	b.ew.Write(b.events[:])
	return persist(ctx, newSession)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer).setup()
	},
}

func (b *Buffer) Save(ctx context.Context) {
	defer b.Release()
	say := log.Get(ctx)
	ts := Get(ctx)
	bob := &Bob{db: ts.db}
	b.id.SetTime(time.Now())
	b.id.Entropy()
	{
		// save events
		b.id.SetTable(EVENTS)
		err := b.ew.Close()
		if err != nil {
			say.Err(err).Msg("failed to close parquet writer for events")
			return
		}
		err = bob.Store2(&b.id, b.eb.Bytes(), b.ttl)
		if err != nil {
			say.Err(err).Msg("failed to save events to permanent storage")
			return
		}
		say.Debug().Msgf("saved  %s events ", units.BytesSize(float64(b.eb.Len())))
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
			say.Err(err).Msg("failed to save events to permanent storage")
			return
		}
		say.Debug().Msgf("saved  %s sessions ", units.BytesSize(float64(b.sb.Len())))
	}
}

func find(ctx context.Context, e *Event, userId int64) *Session {
	v, _ := caches.Session(ctx).Get(key(e.Domain, userId))
	if v != nil {
		return v.(*Session)
	}
	return nil
}

// const of storing a session in cache
var sessionSize = int64(reflect.TypeOf(Session{}).Size())

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

func persist(ctx context.Context, s *Session) uuid.UUID {
	caches.Session(ctx).SetWithTTL(key(s.Domain, s.UserId), s, sessionSize, 30*time.Minute)
	return s.ID
}
