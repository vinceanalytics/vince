package entry

import (
	"sync"
	"time"

	"github.com/polarsignals/frostdb/dynparquet"
	schemapb "github.com/polarsignals/frostdb/gen/proto/go/frostdb/schema/v1alpha2"
	"github.com/segmentio/parquet-go"
)

type Entry struct {
	Browser                string
	BrowserVersion         string
	City                   string
	Country                string
	Domain                 string
	Duration               time.Duration
	EntryPage              string
	ExitPage               string
	Hostname               string
	ID                     uint64
	IsBounce               bool
	Name                   string
	OperatingSystem        string
	OperatingSystemVersion string
	Pathname               string
	Referrer               string
	ReferrerSource         string
	Region                 string
	ScreenSize             string
	Timestamp              int64
	TransferredFrom        string
	UtmCampaign            string
	UtmContent             string
	UtmMedium              string
	UtmSource              string
	UtmTerm                string
	Value                  int64
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
	e.ExitPage = e.Pathname
	e.Value = 1
}

func (s *Entry) Update(e *Entry) {
	s.ExitPage = e.Pathname
	s.Duration = time.UnixMilli(e.Timestamp).Sub(time.UnixMilli(s.Timestamp))
}

func (e *Entry) Row() parquet.Row {
	return parquet.Row{
		boolValue("bounce", e.IsBounce),
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

func boolValue(name string, ok bool) parquet.Value {
	return parquet.BooleanValue(ok).Level(0, 0, columnIndex[name])
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
			nullableString("patch"),
			nullableString("referrer"),
			nullableString("referrer_source"),
			nullableString("region"),
			nullableString("screen"),
			nullableString("utm_campaign"),
			nullableString("utm_content"),
			nullableString("utm_medium"),
			nullableString("utm_source"),
			nullableString("utm_term"),
			plainBool("bounce"),
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

func plainBool(name string) *schemapb.Node {
	return &schemapb.Node{
		Type: &schemapb.Node_Leaf{
			Leaf: &schemapb.Leaf{Name: name, StorageLayout: &schemapb.StorageLayout{
				Type:        schemapb.StorageLayout_TYPE_BOOL,
				Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
			}},
		},
	}
}
