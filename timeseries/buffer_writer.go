package timeseries

import (
	"bytes"
	"context"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/docker/go-units"
	"github.com/gammazero/deque"
	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/log"
	"github.com/google/uuid"
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

func (b *Buffer) Reset() *Buffer {
	for i := range b.id {
		b.id[i] = 0
	}
	b.ttl = 0
	b.eb.Reset()
	b.sb.Reset()
	b.events[0] = nil
	b.sessions[0] = nil
	b.sessions[1] = nil
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
	b.ew = parquet.NewSortingWriter[*Event](&b.eb, SortRowCount,
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		),
		parquet.BloomFilters(eb...),
	)
	sb := make([]parquet.BloomFilterColumn, len(sessionFilterFields))
	for i := range sessionFilterFields {
		sb[i] = parquet.SplitBlockFilter(10, sessionFilterFields[i])
	}
	b.sw = parquet.NewSortingWriter[*Session](&b.sb, SortRowCount,
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		),
		parquet.BloomFilters(sb...),
	)
	return b
}

func NewBuffer(user uint64, ttl time.Duration) *Buffer {
	return bigBufferPool.Get().(*Buffer).Init(user, ttl)
}

func (b *Buffer) Expired() bool {
	return b.start.Add(b.ttl).Before(time.Now())
}

func (b *Buffer) Register(ctx context.Context, e *Event, prevUserId int64) uuid.UUID {
	var s *Session
	s = find(ctx, e, e.UserId)
	if s == nil {
		s = find(ctx, e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = true
		s.Sign = false
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
		err = storeTxn(&b.id, b.eb.Bytes(), b.ttl, txn)
		if err != nil {
			say.Err(err).Msg("failed to save events to permanent storage")
			return err
		}
		say.Debug().Msgf("saved  %s events ", units.BytesSize(float64(b.eb.Len())))
	}
	{
		// save sessions
		b.id.SetTable(SESSIONS)
		err := b.sw.Close()
		if err != nil {
			say.Err(err).Msg("failed to close parquet writer for sessions")
			return err
		}
		err = storeTxn(&b.id, b.sb.Bytes(), b.ttl, txn)
		if err != nil {
			say.Err(err).Msg("failed to save events to permanent storage")
			return err
		}
		say.Debug().Msgf("saved  %s sessions ", units.BytesSize(float64(b.sb.Len())))
	}
	return nil
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
	return s.SessionId
}

type Map struct {
	m    map[uint64]*Buffer
	mu   sync.Mutex
	ring *deque.Deque[*Buffer]
}

func NewMap() *Map {
	return &Map{
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

func (m *Map) Get(ctx context.Context, uid uint64, ttl time.Duration) *Buffer {
	m.mu.Lock()
	defer m.mu.Unlock()
	var add *Buffer
	for {
		if m.ring.Len() == 0 {
			break
		}
		e := m.ring.PopFront()
		if e.Expired() {
			id := e.id.UserID()
			if id == uid {
				e.Save(ctx)
				add = e
			} else {
				e.Save(ctx)
				e.Release()
				delete(m.m, e.id.UserID())
			}
			continue
		}
		m.ring.PushBack(e)
		break
	}
	if add != nil {
		add.Reset().Init(uid, ttl)
		m.ring.PushBack(add)
		return add
	}
	b := NewBuffer(uid, ttl)
	m.m[uid] = b
	m.ring.PushBack(b)
	return b
}
