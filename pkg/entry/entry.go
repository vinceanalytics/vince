package entry

import (
	"io"
	"sync"
	"time"

	"github.com/polarsignals/frostdb/dynparquet"
	schemapb "github.com/polarsignals/frostdb/gen/proto/go/frostdb/schema/v1alpha2"
	"github.com/segmentio/parquet-go"
)

type Entry struct {
	Browser                string        `parquet:"browser,dict,zstd"`
	BrowserVersion         string        `parquet:"browser_version,dict,zstd"`
	City                   string        `parquet:"city,dict,zstd"`
	Country                string        `parquet:"country,dict,zstd"`
	Domain                 string        `parquet:"domain,dict,zstd"`
	Duration               time.Duration `parquet:"duration,dict,zstd"`
	EntryPage              string        `parquet:"entry_page,dict,zstd"`
	ExitPage               string        `parquet:"exit_page,dict,zstd"`
	Hostname               string        `parquet:"host,dict,zstd"`
	ID                     uint64        `parquet:"id,dict,zstd"`
	Bounce                 int64         `parquet:"bounce,dict,zstd"`
	Name                   string        `parquet:"name,dict,zstd"`
	OperatingSystem        string        `parquet:"os,dict,zstd"`
	OperatingSystemVersion string        `parquet:"os_version,dict,zstd"`
	Pathname               string        `parquet:"path,dict,zstd"`
	Referrer               string        `parquet:"referrer,dict,zstd"`
	ReferrerSource         string        `parquet:"referrer_source,dict,zstd"`
	Region                 string        `parquet:"region,dict,zstd"`
	ScreenSize             string        `parquet:"screen,dict,zstd"`
	Timestamp              int64         `parquet:"timestamp,dict,zstd"`
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

func (e *Entry) Release() {
	*e = Entry{}
	entryPool.Put(e)
}

func (e *Entry) Hit() {
	e.EntryPage = e.Pathname
	e.Value = 1
	e.Bounce = 1
}

func (s *Entry) Update(e *Entry) {
	if s.Bounce == 1 {
		s.Bounce, e.Bounce = -1, -1
	}
	e.Value = 1
	e.ExitPage = e.Pathname
	e.Duration = time.UnixMilli(e.Timestamp).Sub(time.UnixMilli(s.Timestamp))
	s.Timestamp = e.Timestamp
}

func (e *Entry) Row() parquet.Row {
	return parquet.Row{
		int64Value("bounce", e.Bounce),
		int64Value("duration", int64(e.Duration)),
		int64Value("id", int64(e.ID)),
		int64Value("timestamp", e.Timestamp),
		int64Value("value", e.Value),
		stringValue("browser", e.Browser),
		stringValue("browser_version", e.BrowserVersion),
		stringValue("city", e.City),
		stringValue("country", e.City),
		stringValue("entry_page", e.EntryPage),
		stringValue("exit_page", e.ExitPage),
		stringValue("host", e.Hostname),
		stringValue("name", e.Name),
		stringValue("os", e.OperatingSystem),
		stringValue("os_version", e.OperatingSystem),
		stringValue("path", e.Pathname),
		stringValue("referrer", e.Referrer),
		stringValue("referrer_source", e.ReferrerSource),
		stringValue("region", e.Region),
		stringValue("screen", e.ScreenSize),
		stringValue("utm_campaign", e.UtmCampaign),
		stringValue("utm_content", e.UtmContent),
		stringValue("utm_medium", e.UtmCampaign),
		stringValue("utm_source", e.UtmSource),
		stringValue("utm_term", e.UtmTerm),
	}
}

func SchemaBuffer() *dynparquet.Buffer {
	return must(schema.NewBufferV2())
}

var schema = must(dynparquet.SchemaFromDefinition(Scheme))

func must[T any](v T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return v
}

var columnIndex = columns()

func columns() (o map[string]int) {
	o = make(map[string]int)
	for i, f := range schema.ParquetSchema().Fields() {
		o[f.Name()] = i
	}
	return
}

func stringValue(name string, v string) parquet.Value {
	if v == "" {
		return parquet.NullValue().Level(0, 0, columnIndex[name])
	}
	return parquet.ByteArrayValue([]byte(v)).Level(0, 0, columnIndex[name])
}

func int64Value(name string, v int64) parquet.Value {
	return parquet.Int64Value(v).Level(0, 0, columnIndex[name])
}

var Scheme = &schemapb.Schema{
	Root: &schemapb.Group{
		Name: "site_stats",
		Nodes: []*schemapb.Node{
			nullableString("browser"),
			nullableString("browser_version"),
			nullableString("city"),
			nullableString("country_code"),
			nullableString("entry_page"),
			nullableString("exit_page"),
			nullableString("host"),
			nullableString("name"),
			nullableString("os"),
			nullableString("os_version"),
			nullableString("path"),
			nullableString("referrer"),
			nullableString("referrer_source"),
			nullableString("region"),
			nullableString("screen"),
			nullableString("utm_campaign"),
			nullableString("utm_content"),
			nullableString("utm_medium"),
			nullableString("utm_source"),
			nullableString("utm_term"),
			plainInt64("bounce"),
			plainInt64("duration"),
			plainInt64("id"),
			plainInt64("timestamp"),
			plainInt64("value"),
		},
	},
	SortingColumns: []*schemapb.SortingColumn{
		{
			Path:      "timestamp",
			Direction: schemapb.SortingColumn_DIRECTION_ASCENDING,
		},
	},
}

func nullableString(name string) *schemapb.Node {
	return &schemapb.Node{
		Type: &schemapb.Node_Leaf{
			Leaf: &schemapb.Leaf{Name: name, StorageLayout: &schemapb.StorageLayout{
				Type:        schemapb.StorageLayout_TYPE_STRING,
				Encoding:    schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
				Nullable:    true,
				Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
			}},
		},
	}
}

func plainInt64(name string) *schemapb.Node {
	return &schemapb.Node{
		Type: &schemapb.Node_Leaf{
			Leaf: &schemapb.Leaf{Name: name, StorageLayout: &schemapb.StorageLayout{
				Type:        schemapb.StorageLayout_TYPE_INT64,
				Encoding:    schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
				Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
			}},
		},
	}
}

func Writer(w io.Writer) *parquet.SortingWriter[Entry] {
	return parquet.NewSortingWriter[Entry](w, 4<<10, parquet.SortingWriterConfig(
		parquet.SortingColumns(
			parquet.Ascending("timestamp"),
		),
	))
}
