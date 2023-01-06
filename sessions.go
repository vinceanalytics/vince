package vince

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

var sessionPool = &sync.Pool{
	New: func() any {
		return &Session{}
	},
}

const MAX_BUFFER_SIZE = 4098

// Buffers Sessions before they are processed. After processing call Reset to
// recycle the buffer.
type SessionBuffer struct {
	sessions []*Session
}

func (s *SessionBuffer) Reset() {
	s.sessions = s.sessions[:0]
	sessionBufferPool.Put(s)
}

func (s *SessionBuffer) Insert(sess ...*Session) bool {
	s.sessions = append(s.sessions, sess...)
	return len(s.sessions) >= MAX_BUFFER_SIZE
}

var sessionBufferPool = &sync.Pool{
	New: func() any {
		return &SessionBuffer{}
	},
}

func (e *Event) NewSession() *Session {
	s := sessionPool.Get().(*Session)
	s.Sign = 1
	s.SessionId = rand.Uint64()
	s.Hostname = e.Hostname
	s.Domain = e.Domain
	s.UserId = e.UserId
	s.EntryPage = e.Pathname
	s.ExitPage = e.Pathname
	s.IsBounce = true
	s.Duration = durationpb.New(0)
	s.PageViews = 0
	if e.Name == "pageview" {
		s.PageViews = 1
	}
	s.Events = 1
	s.Referrer = e.Referrer
	s.ReferrerSource = e.ReferrerSource
	s.UtmMedium = e.UtmMedium
	s.UtmSource = e.UtmSource
	s.UtmCampaign = e.UtmCampaign
	s.UtmContent = e.UtmContent
	s.UtmTerm = e.UtmTerm
	s.CountryCode = e.CountryCode
	s.Subdivision1Code = e.Subdivision1Code
	s.Subdivision2Code = e.Subdivision2Code
	s.CityGeonameId = e.CityGeonameId
	s.ScreenSize = e.ScreenSize
	s.OperatingSystem = e.OperatingSystem
	s.OperatingSystemVersion = e.OperatingSystemVersion
	s.Browser = e.Browser
	s.Timestamp = e.Timestamp
	s.Labels = append(s.Labels, e.Labels...)
	return s
}

func (s *Session) Update(e *Event) *Session {
	ss := proto.Clone(s).(*Session)
	ss.UserId = e.UserId
	ss.Timestamp = e.Timestamp
	ss.ExitPage = e.Pathname
	ss.IsBounce = false
	ss.Duration = durationpb.New(e.Timestamp.AsTime().Sub(ss.Start.AsTime()))
	if e.Name == "pageview" {
		ss.PageViews++
	}
	if ss.CountryCode == "" {
		ss.CountryCode = e.CountryCode
	}
	if ss.Subdivision1Code == "" {
		ss.Subdivision1Code = e.Subdivision1Code
	}
	if ss.Subdivision2Code == "" {
		ss.Subdivision2Code = e.Subdivision2Code
	}
	if ss.CityGeonameId == 0 {
		ss.CityGeonameId = e.CityGeonameId
	}
	if ss.OperatingSystem == "" {
		ss.OperatingSystem = e.OperatingSystem
	}
	if ss.OperatingSystemVersion == "" {
		ss.OperatingSystemVersion = e.OperatingSystemVersion
	}
	if ss.Browser == "" {
		ss.Browser = e.Browser
	}
	if ss.BrowserVersion == "" {
		ss.BrowserVersion = e.BrowserVersion
	}
	if ss.ScreenSize == "" {
		ss.ScreenSize = e.ScreenSize
	}
	ss.Events += 1
	return ss
}

var buffPool = &sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

type SessionCache struct {
	cache   *ristretto.Cache
	process func(*SessionBuffer)
	buf     *SessionBuffer
	mu      sync.Mutex
}

func NewSessionCache(cache *ristretto.Cache, process func(*SessionBuffer)) *SessionCache {
	return &SessionCache{
		cache:   cache,
		process: process,
		buf:     sessionBufferPool.Get().(*SessionBuffer),
	}
}

func (c *SessionCache) OnEvent(e *Event, prevUserId uint64) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	var s *Session
	s = c.Find(e, e.UserId)
	if s == nil {
		s = c.Find(e, prevUserId)
	}
	if s != nil {
		updated := s.Update(e)
		updated.Sign = 1
		s.Sign = -1
		if c.buf.Insert(updated, s) {
			if c.process != nil {
				c.process(c.buf)
			}
			c.buf = sessionBufferPool.Get().(*SessionBuffer)
		}
		return c.Persist(updated)
	}
	newSession := e.NewSession()
	if c.buf.Insert(newSession) {
		if c.process != nil {
			c.process(c.buf)
		}
		c.buf = sessionBufferPool.Get().(*SessionBuffer)
	}
	return c.Persist(newSession)
}

func (c *SessionCache) Find(e *Event, userId uint64) *Session {
	b := buffPool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		buffPool.Put(b)
	}()
	fmt.Fprintf(b, "%s-%d", e.Domain, userId)
	v, _ := c.cache.Get(b.String())
	if v != nil {
		return v.(*Session)
	}
	return nil
}

func (c *SessionCache) Persist(s *Session) uint64 {
	b := buffPool.Get().(*bytes.Buffer)
	defer func() {
		b.Reset()
		buffPool.Put(b)
	}()
	fmt.Fprintf(b, "%s-%d", s.Domain, s.UserId)
	c.cache.SetWithTTL(b.String(), s, 1, 30*time.Minute)
	return s.SessionId
}
