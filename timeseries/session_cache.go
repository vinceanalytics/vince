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
	cache   *ristretto.Cache
	process chan<- *Session
}

func NewSessionCache(cache *ristretto.Cache, process chan<- *Session) *SessionCache {
	return &SessionCache{
		cache:   cache,
		process: process,
	}
}

func (c *SessionCache) RegisterSession(e *Event, prevUserId uint64) uuid.UUID {
	var s *Session
	s = c.Find(e, e.UserId)
	if s == nil {
		s = c.Find(e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		c.process <- updated
		c.process <- s
		return c.Persist(updated)
	}
	newSession := e.NewSession()
	c.process <- newSession
	return c.Persist(newSession)
}

func (c *SessionCache) Find(e *Event, userId uint64) *Session {
	v, _ := c.cache.Get(key(e.Domain, userId))
	if v != nil {
		return v.(*Session)
	}
	return nil
}

// const of storing a session in cache
var sessionSize = reflect.TypeOf(Session{}).Size()

func key(domain string, userId uint64) string {
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
