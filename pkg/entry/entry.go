package entry

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
)

type Entry struct {
	UID, SID               uint64
	UtmMedium              string
	Referrer               string
	Domain                 string
	ExitPage               string
	EntryPage              string
	Hostname               string
	Pathname               string
	UtmSource              string
	ReferrerSource         string
	Country                string
	Region                 string
	TransferredFrom        string
	UtmCampaign            string
	OperatingSystem        string
	Browser                string
	UtmTerm                string
	Name                   string
	ScreenSize             string
	BrowserVersion         string
	OperatingSystemVersion string
	UtmContent             string
	UserId                 uint64
	SessionId              uint64
	Timestamp              int64
	Duration               time.Duration
	Start                  int64
	City                   string
	PageViews              int32
	Events                 int32
	IsBounce               bool
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
	clone := NewEntry()
	*clone = *e
	return clone
}

func (e *Entry) Release() {
	*e = Entry{}
	entryPool.Put(e)
}

// Session creates a new session from entry
func (e *Entry) Session() {
	e.Start = e.Timestamp
	session := uuid.New()
	e.SessionId = xxhash.Sum64(session[:])
	e.EntryPage = e.Pathname
	e.ExitPage = e.Pathname
	e.IsBounce = true
	e.PageViews = 0
	if e.Name == "pageview" {
		e.PageViews = 1
	}
	e.Events = 1
}

func (s *Entry) Update(e *Entry) {
	s.UserId = e.UserId
	s.Timestamp = e.Timestamp
	s.ExitPage = e.Pathname
	s.IsBounce = false
	s.Duration = time.Unix(e.Timestamp, 0).Sub(time.Unix(s.Start, 0))
	if e.Name == "pageview" {
		s.PageViews++
	}
	if s.Country == "" {
		s.Country = e.Country
	}
	if s.City == "" {
		s.City = e.City
	}
	if s.Region == "" {
		s.Region = e.Region
	}
	if s.OperatingSystem == "" {
		s.OperatingSystem = e.OperatingSystem
	}
	if s.OperatingSystemVersion == "" {
		s.OperatingSystemVersion = e.OperatingSystemVersion
	}
	if s.Browser == "" {
		s.Browser = e.Browser
	}
	if s.BrowserVersion == "" {
		s.BrowserVersion = e.BrowserVersion
	}
	if s.ScreenSize == "" {
		s.ScreenSize = e.ScreenSize
	}
	s.Events += 1
}
