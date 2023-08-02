package entry

import (
	"context"
	"sync"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/compute"
	"github.com/apache/arrow/go/v13/arrow/memory"
)

type Entry struct {
	Bounce         int64
	Session        int64
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
	mu             sync.RWMutex
	Bounce         []int64
	Browser        []string
	BrowserVersion []string
	City           []string
	Country        []string
	Duration       []int64
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
	Session        []int64
	Timestamp      []arrow.Timestamp
	UtmCampaign    []string
	UtmContent     []string
	UtmMedium      []string
	UtmSource      []string
	UtmTerm        []string
	build          *array.RecordBuilder
}

var multiPool = &sync.Pool{
	New: func() any {
		return &MultiEntry{
			Bounce:         make([]int64, 0, 1<<10),
			Browser:        make([]string, 0, 1<<10),
			BrowserVersion: make([]string, 0, 1<<10),
			City:           make([]string, 0, 1<<10),
			Country:        make([]string, 0, 1<<10),
			Duration:       make([]int64, 0, 1<<10),
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
			Session:        make([]int64, 0, 1<<10),
			Timestamp:      make([]arrow.Timestamp, 0, 1<<10),
			UtmCampaign:    make([]string, 0, 1<<10),
			UtmContent:     make([]string, 0, 1<<10),
			UtmMedium:      make([]string, 0, 1<<10),
			UtmSource:      make([]string, 0, 1<<10),
			UtmTerm:        make([]string, 0, 1<<10),
			build:          array.NewRecordBuilder(Pool, Schema),
		}
	},
}

func NewMulti() *MultiEntry {
	return multiPool.Get().(*MultiEntry)
}

func (m *MultiEntry) Release() {
	m.Reset()
	multiPool.Put(m)
}

func (m *MultiEntry) Reset() {
	m.Bounce = m.Bounce[:0]
	m.Browser = m.Browser[:0]
	m.BrowserVersion = m.BrowserVersion[:0]
	m.City = m.City[:0]
	m.Country = m.Country[:0]
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
	m.Session = m.Session[:0]
	m.Timestamp = m.Timestamp[:0]
	m.UtmCampaign = m.UtmCampaign[:0]
	m.UtmContent = m.UtmContent[:0]
	m.UtmMedium = m.UtmMedium[:0]
	m.UtmSource = m.UtmSource[:0]
	m.UtmTerm = m.UtmTerm[:0]
}

func (m *MultiEntry) Append(e *Entry) {
	m.mu.Lock()
	m.Bounce = append(m.Bounce, e.Bounce)
	m.Browser = append(m.Browser, e.Browser)
	m.BrowserVersion = append(m.BrowserVersion, e.BrowserVersion)
	m.City = append(m.City, e.City)
	m.Country = append(m.Country, e.Country)
	m.Duration = append(m.Duration, int64(e.Duration))
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
	m.Session = append(m.Session, e.Session)
	m.Timestamp = append(m.Timestamp, 0) // This will be updated when saving
	m.UtmCampaign = append(m.UtmCampaign, e.UtmCampaign)
	m.UtmContent = append(m.UtmContent, e.UtmContent)
	m.UtmMedium = append(m.UtmMedium, e.UtmMedium)
	m.UtmSource = append(m.UtmSource, e.UtmSource)
	m.UtmTerm = append(m.UtmTerm, e.UtmTerm)
	m.mu.Unlock()
}

func (m *MultiEntry) Record(ts int64) arrow.Record {
	a := arrow.Timestamp(ts)
	for i := range m.Timestamp {
		m.Timestamp[i] = a
	}
	m.build.Reserve(len(m.Timestamp))
	m.int64("bounce", m.Bounce)
	m.string("browser", m.Browser)
	m.string("browser_version", m.BrowserVersion)
	m.string("city", m.City)
	m.string("country", m.Country)
	m.int64("duration", m.Duration)
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
	m.int64("session", m.Session)
	m.timestamp("timestamp", m.Timestamp)
	m.string("utm_campaign", m.UtmCampaign)
	m.string("utm_content", m.UtmContent)
	m.string("utm_medium", m.UtmMedium)
	m.string("utm_source", m.UtmSource)
	m.string("utm_term", m.UtmTerm)
	return m.build.NewRecord()
}

func (m *MultiEntry) int64(name string, values []int64) {
	m.build.Field(Index[name]).(*array.Int64Builder).AppendValues(values, nil)
}

func (m *MultiEntry) string(name string, values []string) {
	m.build.Field(Index[name]).(*array.StringBuilder).AppendStringValues(values, nil)
}

func (m *MultiEntry) timestamp(name string, values []arrow.Timestamp) {
	m.build.Field(Index[name]).(*array.TimestampBuilder).AppendValues(values, nil)
}

// Fields for constructing arrow schema on Entry.
func Fields() []arrow.Field {
	return []arrow.Field{
		{Name: "bounce", Type: arrow.PrimitiveTypes.Int64, Metadata: metaData},
		{Name: "browser", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "browser_version", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "city", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "country", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "duration", Type: arrow.PrimitiveTypes.Int64, Metadata: metaData},
		{Name: "entry_page", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "exit_page", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "host", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "id", Type: arrow.PrimitiveTypes.Int64, Metadata: metaData},
		{Name: "name", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "os", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "os_version", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "path", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "referrer", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "referrer_source", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "region", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "screen", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "session", Type: arrow.PrimitiveTypes.Int64, Metadata: metaData},
		{Name: "timestamp", Type: &arrow.TimestampType{Unit: arrow.Millisecond}, Metadata: metaData},
		{Name: "utm_campaign", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "utm_content", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "utm_medium", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "utm_source", Type: arrow.BinaryTypes.String, Metadata: metaData},
		{Name: "utm_term", Type: arrow.BinaryTypes.String, Metadata: metaData},
	}
}

var all = Fields()

var Index = func() (m map[string]int) {
	m = make(map[string]int)
	for i := range all {
		m[all[i].Name] = i
	}
	return
}()

var metaData = arrow.NewMetadata([]string{"PARQUET:field_id"}, []string{"-1"})

var IndexedColumnsNames = []string{
	"browser",
	"browser_version",
	"city",
	"country",
	"entry_page",
	"exit_page",
	"host",
	"name",
	"os",
	"os_version",
	"path",
	"referrer",
	"referrer_source",
	"region",
	"screen",
	"utm_campaign",
	"utm_content",
	"utm_medium",
	"utm_source",
	"utm_term",
}

// Maps column index to column name of all indexed columns
var IndexedColumns = func() (m map[int]string) {
	m = make(map[int]string)
	for _, n := range IndexedColumnsNames {
		m[Index[n]] = n
	}
	return
}()

var Schema = arrow.NewSchema(all, nil)

func Select(names ...string) *arrow.Schema {
	o := make([]arrow.Field, len(names))
	for i := range o {
		o[i] = Schema.Field(Index[names[i]])
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
	e.Session = 1
}

func (s *Entry) Update(e *Entry) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	}
	e.ExitPage = e.Path
	e.Session = 0
	e.Duration = e.Timestamp.Sub(s.Timestamp)
	s.Timestamp = e.Timestamp
}

func Context(ctx ...context.Context) context.Context {
	if len(ctx) > 0 {
		return compute.WithAllocator(ctx[0], Pool)
	}
	return compute.WithAllocator(context.Background(), Pool)
}
