package vince

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/dgraph-io/ristretto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func (e *Event) NewSession() *Session {
	s := GetSession()
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

func (c *SessionCache) RegisterSession(e *Event, prevUserId uint64) uint64 {
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
	b := getBuff()
	defer putBuff(b)
	fmt.Fprintf(b, "%s-%d", e.Domain, userId)
	v, _ := c.cache.Get(b.String())
	if v != nil {
		return v.(*Session)
	}
	return nil
}

// const of storing a session in cache
var sessionSize = reflect.TypeOf(Session{}).Size()

func (c *SessionCache) Persist(s *Session) uint64 {
	b := getBuff()
	defer putBuff(b)
	fmt.Fprintf(b, "%s-%d", s.Domain, s.UserId)
	c.cache.SetWithTTL(b.String(), s, int64(sessionSize), 30*time.Minute)
	return s.SessionId
}
