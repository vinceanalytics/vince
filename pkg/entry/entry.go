package entry

import (
	"sync"
	"time"

	"github.com/apache/arrow/go/v13/arrow"
)

type Entry struct {
	Bounce         int64         `parquet:"bounce,dict,zstd"`
	Browser        string        `parquet:"browser,dict,zstd"`
	BrowserVersion string        `parquet:"browser_version,dict,zstd"`
	City           string        `parquet:"city,dict,zstd"`
	Country        string        `parquet:"country,dict,zstd"`
	Domain         string        `parquet:"domain,dict,zstd"`
	Duration       time.Duration `parquet:"duration,dict,zstd"`
	EntryPage      string        `parquet:"entry_page,dict,zstd"`
	ExitPage       string        `parquet:"exit_page,dict,zstd"`
	Host           string        `parquet:"host,dict,zstd"`
	ID             uint64        `parquet:"id,dict,zstd"`
	Name           string        `parquet:"name,dict,zstd"`
	Os             string        `parquet:"os,dict,zstd"`
	OsVersion      string        `parquet:"os_version,dict,zstd"`
	Path           string        `parquet:"path,dict,zstd"`
	Referrer       string        `parquet:"referrer,dict,zstd"`
	ReferrerSource string        `parquet:"referrer_source,dict,zstd"`
	Region         string        `parquet:"region,dict,zstd"`
	Screen         string        `parquet:"screen,dict,zstd"`
	Timestamp      time.Time     `parquet:"timestamp,timestamp,zstd"`
	UtmCampaign    string        `parquet:"utm_campaign,dict,zstd"`
	UtmContent     string        `parquet:"utm_content,dict,zstd"`
	UtmMedium      string        `parquet:"utm_medium,dict,zstd"`
	UtmSource      string        `parquet:"utm_source,dict,zstd"`
	UtmTerm        string        `parquet:"utm_term,dict,zstd"`
	Value          int64         `parquet:"value,zstd"`
}

// Fields for constructing arrow schema on Entry.
func Fields() []arrow.Field {
	return []arrow.Field{
		{Name: "bounce", Type: arrow.PrimitiveTypes.Int64},
		{Name: "browser", Type: &arrow.StringType{}},
		{Name: "browser_version", Type: &arrow.StringType{}},
		{Name: "city", Type: &arrow.StringType{}},
		{Name: "country", Type: &arrow.StringType{}},
		{Name: "domain", Type: &arrow.StringType{}},
		{Name: "duration", Type: &arrow.DurationType{}},
		{Name: "entry_page", Type: &arrow.StringType{}},
		{Name: "exit_page", Type: &arrow.StringType{}},
		{Name: "host", Type: &arrow.StringType{}},
		{Name: "id", Type: &arrow.StringType{}},
		{Name: "name", Type: &arrow.StringType{}},
		{Name: "os", Type: &arrow.StringType{}},
		{Name: "os_version", Type: &arrow.StringType{}},
		{Name: "path", Type: &arrow.StringType{}},
		{Name: "referrer", Type: &arrow.StringType{}},
		{Name: "referrer_source", Type: &arrow.StringType{}},
		{Name: "region", Type: &arrow.StringType{}},
		{Name: "screen", Type: &arrow.StringType{}},
		{Name: "timestamp", Type: &arrow.TimestampType{Unit: arrow.Millisecond}},
		{Name: "utm_campaign", Type: &arrow.StringType{}},
		{Name: "utm_content", Type: &arrow.StringType{}},
		{Name: "utm_medium", Type: &arrow.StringType{}},
		{Name: "utm_source", Type: &arrow.StringType{}},
		{Name: "utm_term", Type: &arrow.StringType{}},
		{Name: "utm_term", Type: &arrow.StringType{}},
		{Name: "value", Type: arrow.PrimitiveTypes.Int64},
	}
}

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
	e.Value = 1
}

func (s *Entry) Update(e *Entry) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	}
	e.ExitPage = e.Path
	e.Duration = e.Timestamp.Sub(s.Timestamp)
	s.Timestamp = e.Timestamp
}
