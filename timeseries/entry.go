package timeseries

import (
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
)

// Session creates a new session from entry
func (e *Entry) Session() *Entry {
	s := *e
	s.Sign = 1
	s.IsSession = true
	session := uuid.New()
	s.SessionId = xxhash.Sum64(session[:])
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

func (e *Entry) Bounce() (n int32) {
	if e.IsBounce {
		n = 1
	}
	return
}

func (s *Entry) Update(e *Entry) *Entry {
	ss := *s
	ss.UserId = e.UserId
	ss.Timestamp = e.Timestamp
	ss.ExitPage = e.Pathname
	ss.IsBounce = false
	ss.Duration = int64(time.Unix(e.Timestamp, 0).Sub(time.Unix(ss.Start, 0)))
	if e.Name == "pageview" {
		ss.PageViews++
	}
	if ss.CountryCode == "" {
		ss.CountryCode = e.CountryCode
	}
	if ss.CityGeoNameId == "" {
		ss.CityGeoNameId = e.CityGeoNameId
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

type EntryList []*Entry

func (e *Entries) List() EntryList {
	return EntryList(e.Events)
}

func (ls EntryList) Count(u, s *roaring64.Bitmap) (visitors, visits, events float64) {
	if len(ls) == 0 {
		return
	}
	u.Clear()
	s.Clear()
	for _, e := range ls {
		if !e.IsSession {
			events += 1
		}
		if !u.Contains(e.UserId) {
			u.Add(e.UserId)
			visitors += 1
		}
		if !s.Contains(e.SessionId) {
			u.Add(e.SessionId)
			visits += 1
		}
	}
	return
}

// for collects entries happening within an hour and calls f with the hour and the list
// of entries.
//
// Assumes ls is sorted and contains entries happening in the same day.
func (ls EntryList) Emit(f func(EntryList)) {
	var pos int
	var last, curr int64
	for i := range ls {
		curr = ls[i].Timestamp
		if i > 0 && curr != last {
			f(ls[pos : i-1])
			pos = i
		}
		last = curr
	}
	if pos < len(ls)-1 {
		f(ls[pos:])
	}
}
