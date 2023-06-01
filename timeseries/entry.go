package timeseries

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
)

type Entry struct {
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
	Sign                   int32
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
func (e *Entry) Session() *Entry {
	e.Sign = 1
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
	return e
}

func (e *Entry) Bounce() (n int32) {
	if e.IsBounce {
		n = 1
	}
	return
}

func (s *Entry) Update(e *Entry) *Entry {
	ss := NewEntry()
	*ss = *s
	ss.UserId = e.UserId
	ss.Timestamp = e.Timestamp
	ss.ExitPage = e.Pathname
	ss.IsBounce = false
	ss.Duration = time.Unix(e.Timestamp, 0).Sub(time.Unix(ss.Start, 0))
	if e.Name == "pageview" {
		ss.PageViews++
	}
	if ss.Country == "" {
		ss.Country = e.Country
	}
	if ss.City == "" {
		ss.City = e.City
	}
	if ss.Region == "" {
		ss.Region = e.Region
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
