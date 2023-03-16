package timeseries

import (
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v12/arrow"
	"github.com/google/uuid"
	"github.com/segmentio/parquet-go/bloom/xxhash"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Entry represent an event/session with web analytics payload.
type Entry struct {
	Timestamp              time.Time     `parquet:"timestamp"`
	Name                   string        `parquet:"name,dict,zstd"`
	Domain                 string        `parquet:"domain,dict,zstd"`
	UserId                 uint64        `parquet:"user_id,dict,zstd"`
	SessionId              uuid.UUID     `parquet:"session_id,dict,zstd"`
	Hostname               string        `parquet:"hostname,dict,zstd"`
	Pathname               string        `parquet:"path,dict,zstd"`
	Referrer               string        `parquet:"referrer,dict,zstd"`
	ReferrerSource         string        `parquet:"referrer_source,dict,zstd"`
	CountryCode            string        `parquet:"country_code,dict,zstd"`
	Region                 string        `parquet:"region,dict,zstd"`
	ScreenSize             string        `parquet:"screen_size,dict,zstd"`
	OperatingSystem        string        `parquet:"operating_system,dict,zstd"`
	Browser                string        `parquet:"browser,dict,zstd"`
	UtmMedium              string        `parquet:"utm_medium,dict,zstd"`
	UtmSource              string        `parquet:"utm_source,dict,zstd"`
	UtmCampaign            string        `parquet:"utm_campaign,dict,zstd"`
	BrowserVersion         string        `parquet:"browser_version,dict,zstd"`
	OperatingSystemVersion string        `parquet:"operating_system_version,dict,zstd"`
	UtmContent             string        `parquet:"utm_content,dict,zstd"`
	UtmTerm                string        `parquet:"utm_term,dict,zstd"`
	TransferredFrom        string        `parquet:"transferred_from,dict,zstd"`
	EntryPage              string        `parquet:"entry_page,dict,zstd"`
	ExitPage               string        `parquet:"exit_page,dict,zstd"`
	CityGeoNameID          uint32        `parquet:"city_geo_name_id,dict,zstd"`
	PageViews              int64         `parquet:"pageviews,dict,zstd"`
	Events                 int64         `parquet:"events,dict,zstd"`
	Sign                   int32         `parquet:"sign,dict,zstd"`
	IsBounce               bool          `parquet:"is_bounce,dict,zstd"`
	Duration               time.Duration `parquet:"duration,dict,zstd"`
	Start                  time.Time     `parquet:"start,zstd"`

	hash uint64 // xxhash of SessionId
}

func (e *Entry) Hash() uint64 {
	if e.hash != 0 {
		return e.hash
	}
	e.hash = xxhash.Sum64(e.SessionId[:])
	return e.hash
}

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
	s.SessionId = uuid.New()
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

func (s *Entry) Update(e *Entry) *Entry {
	ss := *s
	ss.UserId = e.UserId
	ss.Timestamp = e.Timestamp
	ss.ExitPage = e.Pathname
	ss.IsBounce = false
	ss.Duration = e.Timestamp.Sub(ss.Start)
	if e.Name == "pageview" {
		ss.PageViews++
	}
	if ss.CountryCode == "" {
		ss.CountryCode = e.CountryCode
	}
	if ss.CityGeoNameID == 0 {
		ss.CityGeoNameID = e.CityGeoNameID
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

func (ls EntryList) Aggregate(u, s *roaring64.Bitmap) (a *Aggregate) {
	a = &Aggregate{}
	u.Clear()
	s.Clear()
	var d time.Duration
	var pages int64
	var sign int32
	for _, e := range ls {
		if !u.Contains(e.UserId) {
			u.Add(e.UserId)
			a.Visitors += 1
		}
		if !s.Contains(e.Hash()) {
			u.Add(e.Hash())
			a.Visits += 1
		}
		d += e.Duration
		sign += e.Sign
		pages += e.PageViews
	}
	a.VisitDuration = durationpb.New(d / time.Duration(sign))
	a.ViewsPerVisit = float64(pages) / float64(sign)
	return
}

// for collects entries happening within an hour and calls f with the hour and the list
// of entries.
//
// Assumes ls is sorted and contains entries happening in the same day.
func (ls EntryList) Emit(f func(int, EntryList)) {
	var pos int
	for i := range ls {
		if i > 0 && ls[i].Timestamp.Hour() != ls[i-1].Start.Hour() {
			f(ls[pos].Timestamp.Hour(), ls[pos:i-1])
			pos = i
		}
	}
	if pos < len(ls)-1 {
		f(ls[pos].Timestamp.Hour(), ls[pos:])
	}
}
