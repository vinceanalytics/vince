package timeseries

import (
	"sort"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/gernest/vince/store"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
	"google.golang.org/protobuf/proto"
)

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
	ss := proto.Clone(s).(*Entry)
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
	if ss.CityGeoNameId == 0 {
		ss.CityGeoNameId = e.CityGeoNameId
	}
	if ss.Subdivision1Code == "" {
		ss.Subdivision1Code = e.Subdivision1Code
	}
	if ss.Subdivision2Code == "" {
		ss.Subdivision2Code = e.Subdivision2Code
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

type EntryList []*Entry

func (e *Entries) List() EntryList {
	return EntryList(e.Events)
}

func (ls EntryList) Count(u, s *roaring64.Bitmap, sum *store.Sum) {
	if len(ls) == 0 {
		return
	}
	u.Clear()
	s.Clear()

	return
}

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

func (e EntryList) Sort(by PROPS) {
	switch by {
	case PROPS_NAME:
		sort.Slice(e, func(i, j int) bool {
			return e[i].Name < e[j].Name
		})
	default:
		return
	}
}

func (e EntryList) EmitProp(u, s *roaring64.Bitmap, by PROPS, sum *store.Sum, f func(key string, sum *store.Sum) error) error {
	e.Sort(by)
	var key func(*Entry) string
	switch by {
	case PROPS_NAME:
		key = func(e *Entry) string {
			return e.Name
		}
	case PROPS_UTM_DEVICE:
		key = func(e *Entry) string {
			return e.ScreenSize
		}
	case PROPS_OS:
		key = func(e *Entry) string {
			return e.OperatingSystem
		}
	case PROPS_OS_VERSION:
		key = func(e *Entry) string {
			return e.OperatingSystemVersion
		}
	case PROPS_UTM_BROWSER:
		key = func(e *Entry) string {
			return e.Browser
		}
	case PROPS_BROWSER_VERSION:
		key = func(e *Entry) string {
			return e.BrowserVersion
		}
	case PROPS_REGION:
		key = func(e *Entry) string {
			return e.Subdivision1Code
		}
	case PROPS_COUNTRY:
		key = func(e *Entry) string {
			return e.CountryCode
		}
	default:
		return nil
	}
	var start int
	var lastKey, currentKey string
	for i := range e {
		currentKey = key(e[i])
		if currentKey == "" {
			continue
		}
		if lastKey == "" {
			// empty keys starts first. Here we have non empty key, we start counting
			// for this key forward.
			lastKey = currentKey
			start = i
			continue
		}
		if lastKey != currentKey {
			// we have come across anew key, save the old key
			e[start:i].Count(u, s, sum)
			err := f(lastKey, sum)
			if err != nil {
				return err
			}
			start = i
			lastKey = currentKey
		}
	}
	if start < len(e)-1 {
		e[start:].Count(u, s, sum)
		return f(lastKey, sum)
	}
	return nil
}
