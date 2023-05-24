package timeseries

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
)

type Entry struct {
	Timestamp              int64
	Name                   string
	Domain                 string
	UserId                 uint64
	SessionId              uint64
	Hostname               string
	Pathname               string
	Referrer               string
	ReferrerSource         string
	CountryCode            string
	Subdivision1Code       string
	Subdivision2Code       string
	CityGeoNameId          uint32
	ScreenSize             string
	OperatingSystem        string
	Browser                string
	UtmMedium              string
	UtmSource              string
	UtmCampaign            string
	BrowserVersion         string
	OperatingSystemVersion string
	UtmContent             string
	UtmTerm                string
	TransferredFrom        string
	EntryPage              string
	ExitPage               string
	PageViews              int32
	Events                 int32
	Sign                   int32
	IsBounce               bool
	Duration               float64
	Start                  int64
	HourIndex              int32
}

var entryPool = &sync.Pool{
	New: func() any {
		return new(Entry)
	},
}

func NewEntry() *Entry {
	return entryPool.Get().(*Entry)
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
	ss.Duration = math.Abs(time.Unix(e.Timestamp, 0).Sub(time.Unix(ss.Start, 0)).Seconds())
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

func (ls EntryList) Count(u, s *roaring64.Bitmap, sum *Sum) {
	if len(ls) == 0 {
		return
	}
	*sum = Sum{}
	u.Clear()
	s.Clear()
	var signSum, bounce, views, events, visitors int32
	var duration float64
	for _, e := range ls {
		signSum += e.Sign
		bounce += e.Bounce() * e.Sign
		views += e.PageViews * e.Sign
		events += e.Events * e.Sign
		if !u.Contains(e.UserId) {
			visitors += 1
			u.Add(e.UserId)
		}
		duration += e.Duration * float64(e.Sign)
	}
	sum.BounceRate = uint32(math.Round(float64(bounce) / float64(signSum) * 100))
	sum.Visits = uint32(signSum)
	sum.Views = uint32(views)
	sum.Events = uint32(events)
	sum.Visitors = uint32(visitors)
	sum.VisitDuration = uint32(math.Round(duration / float64(signSum)))
	sum.ViewsPerVisit = uint32(math.Round(float64(views) / float64(signSum)))
}

func (ls EntryList) Emit(f func(EntryList)) {
	if len(ls) < 2 {
		return
	}
	if ls[0].HourIndex == ls[len(ls)-1].HourIndex {
		// ls is stats for the hour. Return early. We don't need to check dates here
		// we know collection windows are short.
		f(ls)
		return
	}
	var pos int
	var last, curr int32
	for i := range ls {
		curr = ls[i].HourIndex
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

func (e EntryList) Sort(by Property) {
	var less func(i, j int) bool
	switch by {
	case Event:
		less = func(i, j int) bool {
			return e[i].Name < e[j].Name
		}
	case Page:
		less = func(i, j int) bool {
			return e[i].Pathname < e[j].Pathname
		}
	case EntryPage:
		less = func(i, j int) bool {
			return e[i].EntryPage < e[j].EntryPage
		}
	case ExitPage:
		less = func(i, j int) bool {
			return e[i].ExitPage < e[j].ExitPage
		}
	case Referrer:
		less = func(i, j int) bool {
			return e[i].Referrer < e[j].Referrer
		}
	case UtmDevice:
		less = func(i, j int) bool {
			return e[i].ScreenSize < e[j].ScreenSize
		}
	case UtmMedium:
		less = func(i, j int) bool {
			return e[i].UtmMedium < e[j].UtmMedium
		}
	case UtmSource:
		less = func(i, j int) bool {
			return e[i].UtmSource < e[j].UtmSource
		}
	case UtmCampaign:
		less = func(i, j int) bool {
			return e[i].UtmCampaign < e[j].UtmCampaign
		}
	case UtmContent:
		less = func(i, j int) bool {
			return e[i].UtmContent < e[j].UtmContent
		}
	case UtmTerm:
		less = func(i, j int) bool {
			return e[i].UtmTerm < e[j].UtmTerm
		}
	case Os:
		less = func(i, j int) bool {
			return e[i].OperatingSystem < e[j].OperatingSystem
		}
	case OsVersion:
		less = func(i, j int) bool {
			return e[i].OperatingSystemVersion < e[j].OperatingSystemVersion
		}
	case UtmBrowser:
		less = func(i, j int) bool {
			return e[i].Browser < e[j].Browser
		}
	case BrowserVersion:
		less = func(i, j int) bool {
			return e[i].BrowserVersion < e[j].BrowserVersion
		}
	case Region:
		less = func(i, j int) bool {
			return e[i].Subdivision1Code < e[j].Subdivision1Code
		}
	case Country:
		less = func(i, j int) bool {
			return e[i].CountryCode < e[j].CountryCode
		}
	case City:
		less = func(i, j int) bool {
			return e[i].CityGeoNameId < e[j].CityGeoNameId
		}
	default:
		return
	}
	sort.Slice(e, less)
}

func (e EntryList) EmitProp(ctx context.Context, cf *saveContext, u, s *roaring64.Bitmap, by Property, sum *Sum, f func(key string, sum *Sum) error) error {
	e.Sort(by)
	var key func(*Entry) string
	switch by {
	case Event:
		key = func(e *Entry) string {
			return e.Name
		}
	case Page:
		key = func(e *Entry) string {
			return e.Pathname
		}
	case EntryPage:
		key = func(e *Entry) string {
			return e.EntryPage
		}
	case ExitPage:
		key = func(e *Entry) string {
			return e.ExitPage
		}
	case Referrer:
		key = func(e *Entry) string {
			return e.Referrer
		}
	case UtmDevice:
		key = func(e *Entry) string {
			return e.ScreenSize
		}
	case UtmMedium:
		key = func(e *Entry) string {
			return e.UtmMedium
		}
	case UtmSource:
		key = func(e *Entry) string {
			return e.UtmSource
		}
	case UtmCampaign:
		key = func(e *Entry) string {
			return e.UtmCampaign
		}
	case UtmContent:
		key = func(e *Entry) string {
			return e.UtmContent
		}
	case UtmTerm:
		key = func(e *Entry) string {
			return e.UtmTerm
		}
	case Os:
		key = func(e *Entry) string {
			return e.OperatingSystem
		}
	case OsVersion:
		key = func(e *Entry) string {
			return e.OperatingSystemVersion
		}
	case UtmBrowser:
		key = func(e *Entry) string {
			return e.Browser
		}
	case BrowserVersion:
		key = func(e *Entry) string {
			return e.BrowserVersion
		}
	case Region:
		key = func(e *Entry) string {
			return e.Subdivision1Code
		}
	case Country:
		key = func(e *Entry) string {
			return e.CountryCode
		}
	case City:
		key = func(e *Entry) string {
			return cf.findCity(ctx, e.CityGeoNameId)
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
			// e is sorted in ascending order. This guarantees empty strings will
			// appear first on the list.
			//
			// Here current key is not empty but last key was empty. We mark this as
			// the start of our iteration.
			lastKey = currentKey
			start = i
			continue
		}
		if lastKey != currentKey {
			// we have come across a new key. Aggregate the old key and call
			// f on the result.
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
