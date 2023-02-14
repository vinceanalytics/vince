package timeseries

import (
	"bytes"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/google/uuid"
)

type SessionCache struct {
	cache    *ristretto.Cache
	sessions chan<- *Session
	events   chan<- *Event
}

func NewSessionCache(cache *ristretto.Cache, sessions chan<- *Session, events chan<- *Event) *SessionCache {
	return &SessionCache{
		cache:    cache,
		sessions: sessions,
		events:   events,
	}
}

func (c *SessionCache) RegisterSession(e *Event, prevUserId int64) uuid.UUID {
	var s *Session
	s = c.Find(e, e.UserId)
	if s == nil {
		s = c.Find(e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		c.sessions <- updated
		c.sessions <- s
		return c.Persist(updated)
	}
	newSession := e.NewSession()
	c.sessions <- newSession
	c.events <- e
	return c.Persist(newSession)
}

func (c *SessionCache) Find(e *Event, userId int64) *Session {
	v, _ := c.cache.Get(key(e.Domain, userId))
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

func (c *SessionCache) Persist(s *Session) uuid.UUID {
	c.cache.SetWithTTL(key(s.Domain, s.UserId), s, int64(sessionSize), 30*time.Minute)
	return s.ID
}

func (c *SessionCache) Close() {
	c.cache.Close()
}
