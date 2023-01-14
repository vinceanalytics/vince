package timeseries

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Name                   string    `parquet:"name"`
	Domain                 string    `parquet:"domain"`
	UserId                 uint64    `parquet:"user_id"`
	SessionId              uuid.UUID `parquet:"session_id"`
	Hostname               string    `parquet:"hostname"`
	Pathname               string    `parquet:"path"`
	Referrer               string    `parquet:"referrer"`
	ReferrerSource         string    `parquet:"referrer_source"`
	CountryCode            string    `parquet:"country_code"`
	ScreenSize             string    `parquet:"screen_size"`
	OperatingSystem        string    `parquet:"operating_system"`
	Browser                string    `parquet:"browser"`
	UtmMedium              string    `parquet:"utm_medium"`
	UtmSource              string    `parquet:"utm_source"`
	UtmCampaign            string    `parquet:"utm_campaign"`
	BrowserVersion         string    `parquet:"browser_version"`
	OperatingSystemVersion string    `parquet:"operating_system_version"`
	CityGeoNameID          uint32    `parquet:"city_geo_name_id"`
	UtmContent             string    `parquet:"utm_content"`
	UtmTerm                string    `parquet:"utm_term"`
	TransferredFrom        string    `parquet:"transferred_from"`
	Labels                 []Label   `parquet:"labels"`
	Timestamp              time.Time `parquet:"timestamp"`
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
	ID                     uuid.UUID     `parquet:"id"`
	Sign                   int32         `parquet:"sign"`
	Domain                 string        `parquet:"domain"`
	UserId                 uint64        `parquet:"user_id"`
	Hostname               string        `parquet:"hostname"`
	IsBounce               bool          `parquet:"is_bounce"`
	EntryPage              string        `parquet:"entry_page"`
	ExitPage               string        `parquet:"exit_page"`
	PageViews              uint64        `parquet:"pageviews"`
	Events                 uint64        `parquet:"events"`
	Duration               time.Duration `parquet:"duration"`
	Referrer               string        `parquet:"referrer"`
	ReferrerSource         string        `parquet:"referrer_source"`
	CountryCode            string        `parquet:"country_code"`
	OperatingSystem        string        `parquet:"operating_system"`
	Browser                string        `parquet:"browser"`
	UtmMedium              string        `parquet:"utm_medium"`
	UtmSource              string        `parquet:"utm_source"`
	UtmCampaign            string        `parquet:"UtmCampaign"`
	BrowserVersion         string        `parquet:"browser_version"`
	OperatingSystemVersion string        `parquet:"operating_system_version"`
	CityGeoNameId          uint32        `parquet:"city_geo_name_id"`
	UtmContent             string        `parquet:"utm_content"`
	UtmTerm                string        `parquet:"utm_term"`
	TransferredFrom        string        `parquet:"transferred_from"`
	ScreenSize             string        `parquet:"screen_size"`
	Labels                 []Label       `parquet:"labels"`
	Start                  time.Time     `parquet:"start"`
	Timestamp              time.Time     `parquet:"timestamp"`
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
