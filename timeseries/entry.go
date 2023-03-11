package timeseries

import (
	"time"

	"github.com/google/uuid"
)

// Entry represent an event/session with web analytics payload.
type Entry struct {
	Timestamp              time.Time     `parquet:"timestamp"`
	Name                   string        `parquet:"name,dict,zstd"`
	Domain                 string        `parquet:"domain,dict,zstd"`
	UserId                 int64         `parquet:"user_id,dict,zstd"`
	SessionId              uuid.UUID     `parquet:"session_id,dict,zstd"`
	Hostname               string        `parquet:"hostname,dict,zstd"`
	Pathname               string        `parquet:"path,dict,zstd"`
	Referrer               string        `parquet:"referrer,dict,zstd"`
	ReferrerSource         string        `parquet:"referrer_source,dict,zstd"`
	CountryCode            string        `parquet:"country_code,dict,zstd"`
	ScreenSize             string        `parquet:"screen_size,dict,zstd"`
	OperatingSystem        string        `parquet:"operating_system,dict,zstd"`
	Browser                string        `parquet:"browser,dict,zstd"`
	UtmMedium              string        `parquet:"utm_medium,dict,zstd"`
	UtmSource              string        `parquet:"utm_source,dict,zstd"`
	UtmCampaign            string        `parquet:"utm_campaign,dict,zstd"`
	BrowserVersion         string        `parquet:"browser_version,dict,zstd"`
	OperatingSystemVersion string        `parquet:"operating_system_version,dict,zstd"`
	CityGeoNameID          uint32        `parquet:"city_geo_name_id,dict,zstd"`
	UtmContent             string        `parquet:"utm_content,dict,zstd"`
	UtmTerm                string        `parquet:"utm_term,dict,zstd"`
	TransferredFrom        string        `parquet:"transferred_from,dict,zstd"`
	Sign                   bool          `parquet:"sign,dict,zstd"`
	IsBounce               bool          `parquet:"is_bounce,dict,zstd"`
	EntryPage              string        `parquet:"entry_page,dict,zstd"`
	ExitPage               string        `parquet:"exit_page,dict,zstd"`
	PageViews              int64         `parquet:"pageviews,dict,zstd"`
	Events                 int64         `parquet:"events,dict,zstd"`
	Duration               time.Duration `parquet:"duration,dict,zstd"`
	Start                  time.Time     `parquet:"start,zstd"`
}

// Session creates a new session from entry
func (e *Entry) Session() *Entry {
	s := *e
	s.Sign = true
	s.SessionId = uuid.New()
	s.EntryPage = e.Pathname
	s.ExitPage = e.Pathname
	s.IsBounce = true
	s.PageViews = 0
	if e.Name == "pageview" {
		s.PageViews = 1
	}
	s.Events = 1
	return &s
}

func (s *Entry) Update(e *Entry) *Entry {
	ss := *s
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
	if ss.CityGeoNameID == 0 {
		ss.CityGeoNameID = e.CityGeoNameID
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
	return &ss
}
