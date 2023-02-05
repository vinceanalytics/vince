package timeseries

import (
	"net/url"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Timestamp              time.Time         `parquet:"timestamp"`
	Name                   string            `parquet:"name,dict,zstd"`
	Domain                 string            `parquet:"domain,dict,zstd"`
	UserId                 uint64            `parquet:"user_id"`
	SessionId              uuid.UUID         `parquet:"session_id,zstd"`
	Hostname               string            `parquet:"hostname,zstd"`
	Pathname               string            `parquet:"path,zstd"`
	Referrer               string            `parquet:"referrer,zstd"`
	ReferrerSource         string            `parquet:"referrer_source,zstd"`
	CountryCode            string            `parquet:"country_code,dict,zstd"`
	ScreenSize             string            `parquet:"screen_size,dict,zstd"`
	OperatingSystem        string            `parquet:"operating_system,dict,zstd"`
	Browser                string            `parquet:"browser,dict,zstd"`
	UtmMedium              string            `parquet:"utm_medium,zstd"`
	UtmSource              string            `parquet:"utm_source,zstd"`
	UtmCampaign            string            `parquet:"utm_campaign,zstd"`
	BrowserVersion         string            `parquet:"browser_version,dict,zstd"`
	OperatingSystemVersion string            `parquet:"operating_system_version,dict,zstd"`
	CityGeoNameID          uint32            `parquet:"city_geo_name_id,dict,zstd"`
	UtmContent             string            `parquet:"utm_content,zstd"`
	UtmTerm                string            `parquet:"utm_term,zstd"`
	TransferredFrom        string            `parquet:"transferred_from,zstd"`
	Labels                 map[string]string `parquet:"labels"`
}

func (e *Event) NewSession() *Session {
	s := new(Session)
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
	if s.Labels == nil {
		s.Labels = make(map[string]string)
	}
	for k, v := range e.Labels {
		s.Labels[k] = v
	}
	return s
}

type Label struct {
	Name  string `parquet:"name,zstd"`
	Value string `parquet:"value,zstd"`
}

type Session struct {
	Timestamp              time.Time         `parquet:"timestamp"`
	ID                     uuid.UUID         `parquet:"id,zstd"`
	Sign                   int32             `parquet:"sign,zstd"`
	Domain                 string            `parquet:"domain,zstd"`
	UserId                 uint64            `parquet:"user_id,zstd"`
	Hostname               string            `parquet:"hostname,zstd"`
	IsBounce               bool              `parquet:"is_bounce,zstd"`
	EntryPage              string            `parquet:"entry_page,zstd"`
	ExitPage               string            `parquet:"exit_page,zstd"`
	PageViews              uint64            `parquet:"pageviews,zstd"`
	Events                 uint64            `parquet:"events,zstd"`
	Duration               time.Duration     `parquet:"duration,zstd"`
	Referrer               string            `parquet:"referrer,zstd"`
	ReferrerSource         string            `parquet:"referrer_source,zstd"`
	CountryCode            string            `parquet:"country_code,zstd"`
	OperatingSystem        string            `parquet:"operating_system,zstd"`
	Browser                string            `parquet:"browser,zstd"`
	UtmMedium              string            `parquet:"utm_medium,zstd"`
	UtmSource              string            `parquet:"utm_source,zstd"`
	UtmCampaign            string            `parquet:"UtmCampaign,zstd"`
	BrowserVersion         string            `parquet:"browser_version,zstd"`
	OperatingSystemVersion string            `parquet:"operating_system_version,zstd"`
	CityGeoNameId          uint32            `parquet:"city_geo_name_id,zstd"`
	UtmContent             string            `parquet:"utm_content,zstd"`
	UtmTerm                string            `parquet:"utm_term,zstd"`
	TransferredFrom        string            `parquet:"transferred_from,zstd"`
	ScreenSize             string            `parquet:"screen_size,zstd"`
	Labels                 map[string]string `parquet:"labels"`
	Start                  time.Time         `parquet:"start"`
}

func (s *Session) Update(e *Event) *Session {
	ss := new(Session)
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

type Tables struct {
	events   *Storage[*Event]
	sessions *Storage[*Session]
}

func Open(dir string) (*Tables, error) {
	base := filepath.Join(dir, "ts")
	events, err := NewStorage[*Event](filepath.Join(base, "events"))
	if err != nil {
		return nil, err
	}
	sessions, err := NewStorage[*Session](filepath.Join(base, "sessions"))
	if err != nil {
		return nil, err
	}
	return &Tables{events: events, sessions: sessions}, nil
}

func (t *Tables) WriteEvents(events []*Event) (int, error) {
	return t.events.Write(events[len(events)-1].Timestamp, events)
}

func (t *Tables) WriteSessions(sessions []*Session) (int, error) {
	return t.sessions.Write(sessions[len(sessions)-1].Timestamp, sessions)
}

func (t *Tables) ArchiveEvents() (int64, error) {
	return t.events.Archive()
}

func (t *Tables) ArchiveSessions() (int64, error) {
	return t.sessions.Archive()
}

func (t *Tables) QueryEvents(params url.Values) (*Record, error) {
	query := QueryFrom(params)
	if query.start.IsZero() {
		return &Record{}, nil
	}
	return t.events.Query(query)
}

func (t *Tables) Close() (err error) {
	if err = t.events.Close(); err != nil {
		return
	}
	err = t.sessions.Close()
	return
}
