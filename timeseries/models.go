package timeseries

import (
	"os"
	"path/filepath"
	"time"

	"github.com/apache/arrow/go/v10/arrow"
	"github.com/apache/arrow/go/v10/arrow/memory"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go"
)

type Event struct {
	Name                   string    `parquet:"name,zstd"`
	Domain                 string    `parquet:"domain,zstd"`
	UserId                 uint64    `parquet:"user_id"`
	SessionId              uuid.UUID `parquet:"session_id,zstd"`
	Hostname               string    `parquet:"hostname,zstd"`
	Pathname               string    `parquet:"path,zstd"`
	Referrer               string    `parquet:"referrer,zstd"`
	ReferrerSource         string    `parquet:"referrer_source,zstd"`
	CountryCode            string    `parquet:"country_code,zstd"`
	ScreenSize             string    `parquet:"screen_size,zstd"`
	OperatingSystem        string    `parquet:"operating_system,zstd"`
	Browser                string    `parquet:"browser,zstd"`
	UtmMedium              string    `parquet:"utm_medium,zstd"`
	UtmSource              string    `parquet:"utm_source,zstd"`
	UtmCampaign            string    `parquet:"utm_campaign,zstd"`
	BrowserVersion         string    `parquet:"browser_version,zstd"`
	OperatingSystemVersion string    `parquet:"operating_system_version,zstd"`
	CityGeoNameID          uint32    `parquet:"city_geo_name_id,zstd"`
	UtmContent             string    `parquet:"utm_content,zstd"`
	UtmTerm                string    `parquet:"utm_term,zstd"`
	TransferredFrom        string    `parquet:"transferred_from,zstd"`
	Labels                 []Label   `parquet:"labels"`
	Timestamp              time.Time `parquet:"timestamp"`
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
	s.Labels = append(s.Labels, e.Labels...)
	return s
}

type Label struct {
	Name  string `parquet:"name,zstd"`
	Value string `parquet:"value,zstd"`
}

type Session struct {
	ID                     uuid.UUID     `parquet:"id,zstd"`
	Sign                   int32         `parquet:"sign,zstd"`
	Domain                 string        `parquet:"domain,zstd"`
	UserId                 uint64        `parquet:"user_id,zstd"`
	Hostname               string        `parquet:"hostname,zstd"`
	IsBounce               bool          `parquet:"is_bounce,zstd"`
	EntryPage              string        `parquet:"entry_page,zstd"`
	ExitPage               string        `parquet:"exit_page,zstd"`
	PageViews              uint64        `parquet:"pageviews,zstd"`
	Events                 uint64        `parquet:"events,zstd"`
	Duration               time.Duration `parquet:"duration,zstd"`
	Referrer               string        `parquet:"referrer,zstd"`
	ReferrerSource         string        `parquet:"referrer_source,zstd"`
	CountryCode            string        `parquet:"country_code,zstd"`
	OperatingSystem        string        `parquet:"operating_system,zstd"`
	Browser                string        `parquet:"browser,zstd"`
	UtmMedium              string        `parquet:"utm_medium,zstd"`
	UtmSource              string        `parquet:"utm_source,zstd"`
	UtmCampaign            string        `parquet:"UtmCampaign,zstd"`
	BrowserVersion         string        `parquet:"browser_version,zstd"`
	OperatingSystemVersion string        `parquet:"operating_system_version,zstd"`
	CityGeoNameId          uint32        `parquet:"city_geo_name_id,zstd"`
	UtmContent             string        `parquet:"utm_content,zstd"`
	UtmTerm                string        `parquet:"utm_term,zstd"`
	TransferredFrom        string        `parquet:"transferred_from,zstd"`
	ScreenSize             string        `parquet:"screen_size,zstd"`
	Labels                 []Label       `parquet:"labels"`
	Start                  time.Time     `parquet:"start"`
	Timestamp              time.Time     `parquet:"timestamp"`
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
	eventsFile     *os.File
	sessionsFile   *os.File
	eventsWriter   *parquet.SortingWriter[*Event]
	sessionsWriter *parquet.SortingWriter[*Session]
	eventsSchema   *arrow.Schema
	sessionsSchema *arrow.Schema
	pool           memory.Allocator
}

func Open(dir string) (*Tables, error) {
	base := filepath.Join(dir, "ts")
	os.MkdirAll(base, 0755)
	eventsDBPath := filepath.Join(base, "events.parquet")
	e, err := os.OpenFile(eventsDBPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	sessionsDBPath := filepath.Join(base, "sessions.parquet")
	s, err := os.OpenFile(sessionsDBPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		e.Close()
		s.Close()
		return nil, err
	}
	t := &Tables{
		eventsFile:   e,
		sessionsFile: s,
		pool:         memory.NewGoAllocator(),
	}
	t.setWriters()
	if err = t.setArrow(); err != nil {
		t.Close()
		return nil, err
	}
	return t, nil
}

func (t *Tables) setWriters() {
	t.eventsWriter = parquet.NewSortingWriter[*Event](
		t.eventsFile,
		4098,
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		),
	)
	t.sessionsWriter = parquet.NewSortingWriter[*Session](
		t.sessionsFile,
		4098,
		parquet.SortingWriterConfig(
			parquet.SortingColumns(
				parquet.Ascending("timestamp"),
			),
		),
	)
}

func (t *Tables) setArrow() error {
	var ea []arrow.Field
	for _, f := range parquet.SchemaOf(&Event{}).Fields() {
		af, err := ParquetFieldToArrowField(f)
		if err != nil {
			return err
		}
		ea = append(ea, af)
	}
	t.eventsSchema = arrow.NewSchema(ea, nil)
	var sa []arrow.Field
	for _, f := range parquet.SchemaOf(&Session{}).Fields() {
		af, err := ParquetFieldToArrowField(f)
		if err != nil {
			return err
		}
		sa = append(sa, af)
	}
	t.sessionsSchema = arrow.NewSchema(sa, nil)
	return nil
}

func (t *Tables) WriteEvents(events []*Event) (int, error) {
	return t.eventsWriter.Write(events)
}

func (t *Tables) WriteSessions(sessions []*Session) (int, error) {
	return t.sessionsWriter.Write(sessions)
}

func (t *Tables) FlushEvents() error {
	return t.eventsWriter.Flush()
}

func (t *Tables) FlushSessions() error {
	return t.sessionsWriter.Flush()
}

func (t *Tables) Close() (err error) {
	if err = t.eventsWriter.Close(); err != nil {
		return
	}
	if err = t.sessionsWriter.Close(); err != nil {
		return
	}
	if err = t.eventsFile.Close(); err != nil {
		return
	}
	if err = t.sessionsFile.Close(); err != nil {
		return
	}
	return
}
