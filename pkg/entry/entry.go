package entry

import (
	"sync"
	"time"
)

type Entry struct {
	Browser                string        `parquet:"browser,dict,zstd"`
	BrowserVersion         string        `parquet:"browser_version,dict,zstd"`
	City                   string        `parquet:"city,dict,zstd"`
	Country                string        `parquet:"country,dict,zstd"`
	Duration               time.Duration `parquet:"duration,dict,zstd"`
	EntryPage              string        `parquet:"entry_page,dict,zstd"`
	ExitPage               string        `parquet:"exit_page,dict,zstd"`
	Hostname               string        `parquet:"host,dict,zstd"`
	ID                     uint64        `parquet:"id,dict,zstd"`
	Bounce                 int64         `parquet:"bounce,dict,zstd"`
	Name                   string        `parquet:"name,dict,zstd"`
	OperatingSystem        string        `parquet:"os,dict,zstd"`
	OperatingSystemVersion string        `parquet:"os_version,dict,zstd"`
	Path                   string        `parquet:"path,dict,zstd"`
	Referrer               string        `parquet:"referrer,dict,zstd"`
	ReferrerSource         string        `parquet:"referrer_source,dict,zstd"`
	Region                 string        `parquet:"region,dict,zstd"`
	ScreenSize             string        `parquet:"screen,dict,zstd"`
	Timestamp              time.Time     `parquet:"timestamp,timestamp,zstd"`
	UtmCampaign            string        `parquet:"utm_campaign,dict,zstd"`
	UtmContent             string        `parquet:"utm_content,dict,zstd"`
	UtmMedium              string        `parquet:"utm_medium,dict,zstd"`
	UtmSource              string        `parquet:"utm_source,dict,zstd"`
	UtmTerm                string        `parquet:"utm_term,dict,zstd"`
	Value                  int64         `parquet:"utm_value,dict,zstd"`
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
	e.Value = 1
	e.Bounce = 1
}

func (s *Entry) Update(e *Entry) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	}
	e.Value = 1
	e.ExitPage = e.Path
	e.Duration = e.Timestamp.Sub(s.Timestamp)
	s.Timestamp = e.Timestamp
}
