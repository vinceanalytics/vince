package timeseries

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Name                   string
	Domain                 string
	UserId                 uint64
	SessionId              uuid.UUID
	Hostname               string
	Pathname               string
	Referrer               string
	ReferrerSource         string
	CountryCode            string
	ScreenSize             string
	OperatingSystem        string
	Browser                string
	UtmMedium              string
	UtmSource              string
	UtmCampaign            string
	BrowserVersion         string
	OperatingSystemVersion string
	CityGeoNameID          uint32
	UtmContent             string
	UtmTerm                string
	TransferredFrom        string
	Labels                 []*Label
	Timestamp              time.Time
}

func (e *Event) NewSession() *Session {
	s := SessionPool.Get().(*Session)
	s.Sign = 1
	s.ID = uuid.New()
	s.Hostname = e.Hostname
	s.Domain = e.Domain
	s.UserId = e.UserId
	s.EntryPage = e.Pathname
	s.ExitPage = e.Pathname
	s.IsBounce = true
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
	s.CityGeoNameId = e.CityGeoNameID
	s.ScreenSize = e.ScreenSize
	s.OperatingSystem = e.OperatingSystem
	s.OperatingSystemVersion = e.OperatingSystemVersion
	s.Browser = e.Browser
	s.Start = e.Timestamp
	s.Timestamp = e.Timestamp
	s.Labels = append(s.Labels, e.Labels...)
	return s
}

func (e *Event) Reset() {
	*e = Event{}
	EventPool.Put(e)
}

type Label struct {
	Name  string
	Value string
}

type Session struct {
	ID                     uuid.UUID
	Sign                   int32
	Domain                 string
	UserId                 uint64
	Hostname               string
	IsBounce               bool
	EntryPage              string
	ExitPage               string
	PageViews              uint64
	Events                 uint64
	Duration               time.Duration
	Referrer               string
	ReferrerSource         string
	CountryCode            string
	OperatingSystem        string
	Browser                string
	UtmMedium              string
	UtmSource              string
	UtmCampaign            string
	BrowserVersion         string
	OperatingSystemVersion string
	CityGeoNameId          uint32
	UtmContent             string
	UtmTerm                string
	TransferredFrom        string
	ScreenSize             string
	Labels                 []*Label
	Start                  time.Time
	Timestamp              time.Time
}

func (s *Session) Reset() {
	*s = Session{}
	SessionPool.Put(s)
}

func (s *Session) Update(e *Event) *Session {
	ss := SessionPool.Get().(*Session)
	*ss = *s
	ss.UserId = e.UserId
	ss.Timestamp = e.Timestamp
	ss.ExitPage = e.Pathname
	ss.IsBounce = false
	ss.Duration = e.Timestamp.Sub(ss.Start)
	if e.Name == "pageview" {
		ss.PageViews++
	}
	if ss.CountryCode == "" {
		ss.CountryCode = e.CountryCode
	}
	if ss.CityGeoNameId == 0 {
		ss.CityGeoNameId = e.CityGeoNameID
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

var SessionPool = &sync.Pool{
	New: func() any {
		return &Session{}
	},
}

var EventPool = &sync.Pool{
	New: func() any {
		return &Event{}
	},
}

func GetEvent() *Event {
	return EventPool.Get().(*Event)
}
