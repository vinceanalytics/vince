package timeseries

import (
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v12/arrow"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
	"google.golang.org/protobuf/types/known/durationpb"
)

// A list of Entry properties as arrow.Field.
var Fields = []arrow.Field{
	{Name: "timestamp", Type: &arrow.TimestampType{Unit: arrow.Nanosecond}},
	{Name: "name", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "domain", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "user_id", Type: &arrow.DictionaryType{IndexType: &arrow.Int64Type{}, ValueType: &arrow.Int64Type{}}},
	{Name: "session_id", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.FixedSizeBinaryType{ByteWidth: 16}}},
	{Name: "hostname", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "path", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "referrer", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "referrer_source", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "country_code", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "screen_size", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "operating_system", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "browser", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "utm_medium", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "utm_source", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "utm_campaign", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "browser_version", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "operating_system_version", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "utm_content", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "utm_term", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "transferred_from", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "entry_page", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "exit_page", Type: &arrow.DictionaryType{IndexType: &arrow.Int32Type{}, ValueType: &arrow.StringType{}}},
	{Name: "city_geo_name_id", Type: &arrow.Int32Type{}},
	{Name: "pageviews", Type: &arrow.Int64Type{}},
	{Name: "events", Type: &arrow.Int64Type{}},
	{Name: "sign", Type: &arrow.Int32Type{}},
	{Name: "is_bounce", Type: &arrow.BooleanType{}},
	{Name: "duration", Type: &arrow.DurationType{}},
	{Name: "start", Type: &arrow.TimestampType{Unit: arrow.Nanosecond}},
}

// Session creates a new session from entry
func (e *Entry) Session() *Entry {
	s := *e
	s.Sign = 1
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

func (ls EntryList) Aggregate(u, s *roaring64.Bitmap) (a *Aggregate) {
	a = &Aggregate{}
	u.Clear()
	s.Clear()
	var d int64
	var pages uint64
	var sign int32
	var bounce int32
	for _, e := range ls {
		if !u.Contains(e.UserId) {
			u.Add(e.UserId)
			a.Visitors += 1
		}
		if !s.Contains(e.SessionId) {
			u.Add(e.SessionId)
			a.Visits += 1
		}
		d += e.Duration
		sign += e.Sign
		pages += e.PageViews
		bounce += e.Sign * e.Bounce()
	}
	a.VisitDuration = durationpb.New(time.Duration(d / int64(sign)))
	a.ViewsPerVisit = float64(pages) / float64(sign)
	bounceRate := (float64(bounce) / float64(sign)) * 100
	a.BounceRate = uint32(bounceRate)
	return
}

// for collects entries happening within an hour and calls f with the hour and the list
// of entries.
//
// Assumes ls is sorted and contains entries happening in the same day.
func (ls EntryList) Emit(f func(int, EntryList)) {
	var pos int
	var last, curr int
	for i := range ls {
		curr = hour(ls[i].Timestamp)
		if i > 0 && curr != last {
			f(curr, ls[pos:i-1])
			pos = i
		}
		last = curr
	}
	if pos < len(ls)-1 {
		f(curr, ls[pos:])
	}
}

func hour(ts int64) int {
	return time.Unix(ts, 0).Hour()
}
