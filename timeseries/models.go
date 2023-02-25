package timeseries

import (
	"context"
	"path/filepath"
	"time"

	"github.com/apache/arrow/go/v12/arrow/memory"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/gernest/vince/log"
	"github.com/google/uuid"
)

const MAX_BUFFER_SIZE = 4098

type Event struct {
	Timestamp              time.Time         `parquet:"timestamp"`
	Name                   string            `parquet:"name,dict,zstd"`
	Domain                 string            `parquet:"domain,dict,zstd"`
	UserId                 int64             `parquet:"user_id"`
	SessionId              uuid.UUID         `parquet:"session_id,zstd"`
	Hostname               string            `parquet:"hostname,zstd"`
	Pathname               string            `parquet:"path,zstd"`
	Referrer               string            `parquet:"referrer,dict,zstd"`
	ReferrerSource         string            `parquet:"referrer_source,dict,zstd"`
	CountryCode            string            `parquet:"country_code,dict,zstd"`
	ScreenSize             string            `parquet:"screen_size,dict,zstd"`
	OperatingSystem        string            `parquet:"operating_system,dict,zstd"`
	Browser                string            `parquet:"browser,dict,zstd"`
	UtmMedium              string            `parquet:"utm_medium,dict,zstd"`
	UtmSource              string            `parquet:"utm_source,dict,zstd"`
	UtmCampaign            string            `parquet:"utm_campaign,dict,zstd"`
	BrowserVersion         string            `parquet:"browser_version,dict,zstd"`
	OperatingSystemVersion string            `parquet:"operating_system_version,dict,zstd"`
	CityGeoNameID          uint32            `parquet:"city_geo_name_id,dict,zstd"`
	UtmContent             string            `parquet:"utm_content,dict,zstd"`
	UtmTerm                string            `parquet:"utm_term,dict,zstd"`
	TransferredFrom        string            `parquet:"transferred_from,dict,zstd"`
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
	Timestamp              time.Time         `parquet:"timestamp,zstd"`
	ID                     uuid.UUID         `parquet:"id,zstd"`
	Sign                   int32             `parquet:"sign,zstd"`
	Domain                 string            `parquet:"domain,dict,zstd"`
	UserId                 int64             `parquet:"user_id,zstd"`
	Hostname               string            `parquet:"hostname,dict,zstd"`
	IsBounce               bool              `parquet:"is_bounce,dict,zstd"`
	EntryPage              string            `parquet:"entry_page,dict,zstd"`
	ExitPage               string            `parquet:"exit_page,dict,zstd"`
	PageViews              int64             `parquet:"pageviews,dict,zstd"`
	Events                 int64             `parquet:"events,dict,zstd"`
	Duration               time.Duration     `parquet:"duration,dict,zstd"`
	Referrer               string            `parquet:"referrer,dict,zstd"`
	ReferrerSource         string            `parquet:"referrer_source,dict,zstd"`
	CountryCode            string            `parquet:"country_code,dict,zstd"`
	OperatingSystem        string            `parquet:"operating_system,dict,zstd"`
	Browser                string            `parquet:"browser,dict,zstd"`
	UtmMedium              string            `parquet:"utm_medium,dict,zstd"`
	UtmSource              string            `parquet:"utm_source,dict,zstd"`
	UtmCampaign            string            `parquet:"UtmCampaign,dict,zstd"`
	BrowserVersion         string            `parquet:"browser_version,dict,zstd"`
	OperatingSystemVersion string            `parquet:"operating_system_version,dict,zstd"`
	CityGeoNameId          uint32            `parquet:"city_geo_name_id,dict,zstd"`
	UtmContent             string            `parquet:"utm_content,dict,zstd"`
	UtmTerm                string            `parquet:"utm_term,dict,zstd"`
	TransferredFrom        string            `parquet:"transferred_from,dict,zstd"`
	ScreenSize             string            `parquet:"screen_size,dict,zstd"`
	Labels                 map[string]string `parquet:"labels"`
	Start                  time.Time         `parquet:"start,zstd"`
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
	db       *badger.DB
	Events   *Storage[*Event]
	Sessions *Storage[*Session]
	Ingest   *Ingest
}

type Ingest struct {
	Events   chan *Event
	Sessions chan *Session
}

func Open(ctx context.Context, allocator memory.Allocator, dir string, ttl time.Duration) (*Tables, error) {
	base := filepath.Join(dir, "ts")
	o := badger.DefaultOptions(filepath.Join(base, "store")).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	db, err := badger.Open(o)
	if err != nil {
		return nil, err
	}

	events, err := NewStorage[*Event](ctx, allocator, db, filepath.Join(base, "events"), ttl)
	if err != nil {
		return nil, err
	}
	sessions, err := NewStorage[*Session](ctx, allocator, db, filepath.Join(base, "sessions"), ttl)
	if err != nil {
		return nil, err
	}
	return &Tables{db: db,
		Events:   events,
		Sessions: sessions,
		Ingest: &Ingest{
			Events:   make(chan *Event, MAX_BUFFER_SIZE),
			Sessions: make(chan *Session, MAX_BUFFER_SIZE),
		},
	}, nil
}

func (t *Tables) WriteEvents(events []*Event) (int, error) {
	return t.Events.Write(events[len(events)-1].Timestamp, events)
}

func (t *Tables) WriteSessions(sessions []*Session) (int, error) {
	return t.Sessions.Write(sessions[len(sessions)-1].Timestamp, sessions)
}

func (t *Tables) ArchiveEvents() (int64, error) {
	return t.Events.Archive()
}

func (t *Tables) ArchiveSessions() (int64, error) {
	return t.Sessions.Archive()
}

func (t *Tables) Close() (err error) {
	t.Events.archive(false)
	t.Events.archive(false)
	err = t.db.Close()
	close(t.Ingest.Events)
	close(t.Ingest.Sessions)
	return
}

type tablesKey struct{}

func Set(ctx context.Context, t *Tables) context.Context {
	return context.WithValue(ctx, tablesKey{}, t)
}

func Get(ctx context.Context) *Tables {
	return ctx.Value(tablesKey{}).(*Tables)
}
