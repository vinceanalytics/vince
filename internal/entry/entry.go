package entry

import (
	"context"
	"io"
	"time"

	"github.com/apache/arrow/go/v14/arrow/compute"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/parquet-go/parquet-go"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
)

type Entry struct {
	Bounce         int64         `parquet:"bounce,dict,zstd"`
	Session        int64         `parquet:"session,dict,zstd"`
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
	Event          string        `parquet:"event,dict,zstd"`
	Os             string        `parquet:"os,dict,zstd"`
	OsVersion      string        `parquet:"os_version,dict,zstd"`
	Path           string        `parquet:"path,dict,zstd"`
	Referrer       string        `parquet:"referrer,dict,zstd"`
	ReferrerSource string        `parquet:"referrer_source,dict,zstd"`
	Region         string        `parquet:"region,dict,zstd"`
	Screen         string        `parquet:"screen,dict,zstd"`
	Timestamp      time.Time     `parquet:"timestamp,timestamp,dict,zstd"`
	UtmCampaign    string        `parquet:"utm_campaign,dict,zstd"`
	UtmContent     string        `parquet:"utm_content,dict,zstd"`
	UtmMedium      string        `parquet:"utm_medium,dict,zstd"`
	UtmSource      string        `parquet:"utm_source,dict,zstd"`
	UtmTerm        string        `parquet:"utm_term,dict,zstd"`
}

var Pool = memory.NewGoAllocator()

func (e *Entry) Clone() *Entry {
	o := *e
	return &o
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

var Schema = parquet.SchemaOf(Entry{})

func NewWriter(input io.Writer) *parquet.GenericWriter[*Entry] {
	var bloom []parquet.BloomFilterColumn
	for i := v1.Column_bounce; i <= v1.Column_utm_term; i++ {
		bloom = append(bloom, parquet.SplitBlockFilter(10, i.String()))
	}
	return parquet.NewGenericWriter[*Entry](input,
		parquet.BloomFilters(bloom...),
	)
}
