package timeseries

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/gernest/vince/caches"
	"github.com/google/uuid"
)

type Buffer struct {
	sessions []*Session
	events   []*Event
}

func (b *Buffer) Reset() {
	b.events = b.events[:0]
	b.sessions = b.sessions[:0]
	bigBufferPool.Put(b)
}

func NewBuffer() *Buffer {
	return bigBufferPool.Get().(*Buffer)
}

var bigBufferPool = &sync.Pool{
	New: func() any {
		return new(Buffer)
	},
}

func RegisterSession(ctx context.Context, e *Event, prevUserId int64) uuid.UUID {
	tab := Get(ctx)
	var s *Session
	s = find(ctx, e, e.UserId)
	if s == nil {
		s = find(ctx, e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		tab.Ingest.Sessions <- updated
		tab.Ingest.Sessions <- s
		return persist(ctx, updated)
	}
	newSession := e.NewSession()
	tab.Ingest.Sessions <- newSession
	tab.Ingest.Events <- e
	return persist(ctx, newSession)
}

func find(ctx context.Context, e *Event, userId int64) *Session {
	v, _ := caches.GetSession(ctx).Get(key(e.Domain, userId))
	if v != nil {
		return v.(*Session)
	}
	return nil
}

// const of storing a session in cache
var sessionSize = reflect.TypeOf(Session{}).Size()

func key(domain string, userId int64) string {
	b := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(b)
	b.Reset()
	fmt.Fprintf(b, "%s%d", domain, userId)
	return b.String()
}

var bufPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func persist(ctx context.Context, s *Session) uuid.UUID {
	caches.GetSession(ctx).SetWithTTL(key(s.Domain, s.UserId), s, int64(sessionSize), 30*time.Minute)
	return s.ID
}
