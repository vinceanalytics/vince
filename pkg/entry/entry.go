package entry

import (
	"sync"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/memory"
)

type Entry struct {
	Bounce         int64
	Browser        string
	BrowserVersion string
	City           string
	Country        string
	Domain         string
	Duration       time.Duration
	EntryPage      string
	ExitPage       string
	Host           string
	ID             uint64
	Name           string
	Os             string
	OsVersion      string
	Path           string
	Referrer       string
	ReferrerSource string
	Region         string
	Screen         string
	Timestamp      time.Time
	UtmCampaign    string
	UtmContent     string
	UtmMedium      string
	UtmSource      string
	UtmTerm        string
}

type MultiEntry struct {
	Bounce         []int64
	Browser        []string
	BrowserVersion []string
	City           []string
	Country        []string
	Domain         []string
	Duration       []arrow.Duration
	EntryPage      []string
	ExitPage       []string
	Host           []string
	ID             []int64
	Name           []string
	Os             []string
	OsVersion      []string
	Path           []string
	Referrer       []string
	ReferrerSource []string
	Region         []string
	Screen         []string
	Timestamp      []arrow.Timestamp
	UtmCampaign    []string
	UtmContent     []string
	UtmMedium      []string
	UtmSource      []string
	UtmTerm        []string
	build          *array.RecordBuilder
}

func NewMulti() *MultiEntry {
	return &MultiEntry{
		Bounce:         make([]int64, 0, 1<<10),
		Browser:        make([]string, 0, 1<<10),
		BrowserVersion: make([]string, 0, 1<<10),
		City:           make([]string, 0, 1<<10),
		Country:        make([]string, 0, 1<<10),
		Domain:         make([]string, 0, 1<<10),
		Duration:       make([]arrow.Duration, 0, 1<<10),
		EntryPage:      make([]string, 0, 1<<10),
		ExitPage:       make([]string, 0, 1<<10),
		Host:           make([]string, 0, 1<<10),
		ID:             make([]int64, 0, 1<<10),
		Name:           make([]string, 0, 1<<10),
		Os:             make([]string, 0, 1<<10),
		OsVersion:      make([]string, 0, 1<<10),
		Path:           make([]string, 0, 1<<10),
		Referrer:       make([]string, 0, 1<<10),
		ReferrerSource: make([]string, 0, 1<<10),
		Region:         make([]string, 0, 1<<10),
		Screen:         make([]string, 0, 1<<10),
		Timestamp:      make([]arrow.Timestamp, 0, 1<<10),
		UtmCampaign:    make([]string, 0, 1<<10),
		UtmContent:     make([]string, 0, 1<<10),
		UtmMedium:      make([]string, 0, 1<<10),
		UtmSource:      make([]string, 0, 1<<10),
		UtmTerm:        make([]string, 0, 1<<10),
		build:          array.NewRecordBuilder(Pool, Schema),
	}
}

func (m *MultiEntry) SetTime(ts time.Time) {
	a := arrow.Timestamp(ts.UnixMilli())
	for i := range m.Timestamp {
		m.Timestamp[i] = a
	}
}

func (m *MultiEntry) Reset() {
	m.Bounce = m.Bounce[:0]
	m.Browser = m.Browser[:0]
	m.BrowserVersion = m.BrowserVersion[:0]
	m.City = m.City[:0]
	m.Country = m.Country[:0]
	m.Domain = m.Domain[:0]
	m.Duration = m.Duration[:0]
	m.EntryPage = m.EntryPage[:0]
	m.ExitPage = m.ExitPage[:0]
	m.Host = m.Host[:0]
	m.ID = m.ID[:0]
	m.Name = m.Name[:0]
	m.Os = m.Os[:0]
	m.OsVersion = m.OsVersion[:0]
	m.Path = m.Path[:0]
	m.Referrer = m.Referrer[:0]
	m.ReferrerSource = m.ReferrerSource[:0]
	m.Region = m.Region[:0]
	m.Screen = m.Screen[:0]
	m.Timestamp = m.Timestamp[:0]
	m.UtmCampaign = m.UtmCampaign[:0]
	m.UtmContent = m.UtmContent[:0]
	m.UtmMedium = m.UtmMedium[:0]
	m.UtmSource = m.UtmSource[:0]
	m.UtmTerm = m.UtmTerm[:0]
}
func (m *MultiEntry) Append(e *Entry) {
	m.Bounce = append(m.Bounce, e.Bounce)
	m.Browser = append(m.Browser, e.Browser)
	m.BrowserVersion = append(m.BrowserVersion, e.BrowserVersion)
	m.City = append(m.City, e.City)
	m.Country = append(m.Country, e.Country)
	m.Domain = append(m.Domain, e.Domain)
	m.Duration = append(m.Duration, arrow.Duration(e.Duration))
	m.EntryPage = append(m.EntryPage, e.EntryPage)
	m.ExitPage = append(m.ExitPage, e.ExitPage)
	m.Host = append(m.Host, e.Host)
	m.ID = append(m.ID, int64(e.ID))
	m.Name = append(m.Name, e.Name)
	m.Os = append(m.Os, e.Os)
	m.OsVersion = append(m.OsVersion, e.OsVersion)
	m.Path = append(m.Path, e.Path)
	m.Referrer = append(m.Referrer, e.Referrer)
	m.ReferrerSource = append(m.ReferrerSource, e.ReferrerSource)
	m.Region = append(m.Region, e.Region)
	m.Screen = append(m.Screen, e.Screen)
	m.Timestamp = append(m.Timestamp, arrow.Timestamp(e.Timestamp.UnixMilli()))
	m.UtmCampaign = append(m.UtmCampaign, e.UtmCampaign)
	m.UtmContent = append(m.UtmContent, e.UtmContent)
	m.UtmMedium = append(m.UtmMedium, e.UtmMedium)
	m.UtmSource = append(m.UtmSource, e.UtmSource)
	m.UtmTerm = append(m.UtmTerm, e.UtmTerm)
}

func (m *MultiEntry) Record() arrow.Record {
	m.build.Reserve(len(m.Timestamp))
	m.int64("bounce", m.Bounce)
	m.string("browser", m.Browser)
	m.string("browser_version", m.BrowserVersion)
	m.string("city", m.City)
	m.string("domain", m.Domain)
	m.duration("duration", m.Duration)
	m.string("entry_page", m.EntryPage)
	m.string("exit_page", m.ExitPage)
	m.string("host", m.Host)
	m.int64("id", m.ID)
	m.string("name", m.Name)
	m.string("os", m.Os)
	m.string("os_version", m.OsVersion)
	m.string("path", m.Path)
	m.string("referrer", m.Referrer)
	m.string("referrer_source", m.ReferrerSource)
	m.string("region", m.Region)
	m.string("screen", m.Screen)
	m.timestamp("timestamp", m.Timestamp)
	m.string("utm_campaign", m.UtmCampaign)
	m.string("utm_content", m.UtmContent)
	m.string("utm_medium", m.UtmMedium)
	m.string("utm_source", m.UtmSource)
	m.string("utm_term", m.UtmTerm)
	return m.build.NewRecord()
}

func (m *MultiEntry) int64(name string, values []int64) {
	m.build.Field(idx[name]).(*array.Int64Builder).AppendValues(values, nil)
}

func (m *MultiEntry) string(name string, values []string) {
	m.build.Field(idx[name]).(*array.StringBuilder).AppendStringValues(values, nil)
}
func (m *MultiEntry) duration(name string, values []arrow.Duration) {
	m.build.Field(idx[name]).(*array.DurationBuilder).AppendValues(values, nil)
}

func (m *MultiEntry) timestamp(name string, values []arrow.Timestamp) {
	m.build.Field(idx[name]).(*array.TimestampBuilder).AppendValues(values, nil)
}

// Fields for constructing arrow schema on Entry.
func Fields() []arrow.Field {
	return []arrow.Field{
		{Name: "bounce", Type: arrow.PrimitiveTypes.Int64},
		{Name: "browser", Type: arrow.BinaryTypes.String},
		{Name: "browser_version", Type: arrow.BinaryTypes.String},
		{Name: "city", Type: arrow.BinaryTypes.String},
		{Name: "country", Type: arrow.BinaryTypes.String},
		{Name: "domain", Type: arrow.BinaryTypes.String},
		{Name: "duration", Type: &arrow.DurationType{}},
		{Name: "entry_page", Type: arrow.BinaryTypes.String},
		{Name: "exit_page", Type: arrow.BinaryTypes.String},
		{Name: "host", Type: arrow.BinaryTypes.String},
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "name", Type: arrow.BinaryTypes.String},
		{Name: "os", Type: arrow.BinaryTypes.String},
		{Name: "os_version", Type: arrow.BinaryTypes.String},
		{Name: "path", Type: arrow.BinaryTypes.String},
		{Name: "referrer", Type: arrow.BinaryTypes.String},
		{Name: "referrer_source", Type: arrow.BinaryTypes.String},
		{Name: "region", Type: arrow.BinaryTypes.String},
		{Name: "screen", Type: arrow.BinaryTypes.String},
		{Name: "timestamp", Type: &arrow.TimestampType{Unit: arrow.Millisecond}},
		{Name: "utm_campaign", Type: arrow.BinaryTypes.String},
		{Name: "utm_content", Type: arrow.BinaryTypes.String},
		{Name: "utm_medium", Type: arrow.BinaryTypes.String},
		{Name: "utm_source", Type: arrow.BinaryTypes.String},
		{Name: "utm_term", Type: arrow.BinaryTypes.String},
		{Name: "utm_term", Type: arrow.BinaryTypes.String},
		{Name: "value", Type: arrow.PrimitiveTypes.Int64},
	}
}

var all = Fields()

var idx = func() (m map[string]int) {
	m = make(map[string]int)
	for i := range all {
		m[all[i].Name] = i
	}
	return
}()

var Schema = arrow.NewSchema(Fields(), nil)

func Select(names ...string) *arrow.Schema {
	o := make([]arrow.Field, len(names))
	for i := range o {
		o[i] = Schema.Field(idx[names[i]])
	}
	return arrow.NewSchema(o, nil)
}

var Pool = memory.NewGoAllocator()

var entryPool = &sync.Pool{
	New: func() any {
		return new(Entry)
	},
}

func NewEntry() *Entry {
	return entryPool.Get().(*Entry)
}

func (e *Entry) Clone() *Entry {
	o := NewEntry()
	*o = *e
	return o
}

func (e *Entry) Release() {
	*e = Entry{}
	entryPool.Put(e)
}

func (e *Entry) Hit() {
	e.EntryPage = e.Path
	e.Bounce = 1
}

func (s *Entry) Update(e *Entry) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	}
	e.ExitPage = e.Path
	e.Duration = e.Timestamp.Sub(s.Timestamp)
	s.Timestamp = e.Timestamp
}
