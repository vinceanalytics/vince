package entry

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/RoaringBitmap/roaring/roaring64"
	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/compute"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/compress"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/schema"
	"github.com/cespare/xxhash/v2"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"golang.org/x/exp/slices"
)

type Entry struct {
	Bounce         int64
	Session        int64
	Browser        string
	BrowserVersion string
	City           string
	Country        string
	Domain         string
	Duration       time.Duration
	EntryPage      string
	ExitPage       string
	Host           string
	ID             uint64
	Event          string
	Os             string
	OsVersion      string
	Path           string
	Referrer       string
	ReferrerSource string
	Region         string
	Screen         string
	Timestamp      int64
	UtmCampaign    string
	UtmContent     string
	UtmMedium      string
	UtmSource      string
	UtmTerm        string
}

type ByteArray struct {
	pos int
	buf []parquet.ByteArray
}

func NewByteArray() ByteArray {
	return ByteArray{
		buf: make([]parquet.ByteArray, 128),
	}
}

func (b *ByteArray) Append(s string) {
	b.grow()
	b.buf[b.pos] = append(b.buf[b.pos], []byte(s)...)
	b.pos++
}

func (b *ByteArray) grow() {
	if b.pos < len(b.buf) {
		return
	}
	b.buf = slices.Grow(b.buf, len(b.buf)*2)
	b.buf = b.buf[:cap(b.buf)]
}

func (b *ByteArray) Reset() {
	for i := range b.buf {
		b.buf[i] = b.buf[i][:0]
	}
}

func (b *ByteArray) Write(g file.ColumnChunkWriter, r *roaring64.Bitmap, field v1.Block_Index_Column) {
	w := g.(*file.ByteArrayColumnChunkWriter)
	must.Must(w.WriteBatch(b.buf[:b.pos], nil, nil))(
		"failed writing int64 column to parquet",
	)
	w.Close()
	if r == nil {
		return
	}
	h := xxhash.New()
	column := []byte(field.String())
	for i := range b.buf[:b.pos] {
		if len(b.buf[i]) == 0 {
			continue
		}
		h.Reset()
		h.Write(column)
		h.Write(b.buf[i])
		sum := h.Sum64()
		if !r.Contains(sum) {
			r.Add(sum)
		}
	}
}

type Int64Array struct {
	buf []int64
}

func NewInt64Array() Int64Array {
	return Int64Array{
		buf: make([]int64, 0, 128),
	}
}

func (b *Int64Array) First() int64 {
	if len(b.buf) == 0 {
		return 0
	}
	return b.buf[0]
}

func (b *Int64Array) Last() int64 {
	if len(b.buf) == 0 {
		return 0
	}
	return b.buf[len(b.buf)-1]
}

func (b *Int64Array) Append(v int64) {
	b.buf = append(b.buf, v)
}

func (b *Int64Array) Reset() {
	b.buf = b.buf[:0]
}

func (b *Int64Array) Write(g file.ColumnChunkWriter) {
	w := g.(*file.Int64ColumnChunkWriter)
	must.Must(w.WriteBatch(b.buf, nil, nil))(
		"failed writing int64 column to parquet",
	)
	w.Close()
}

type MultiEntry struct {
	mu             sync.RWMutex
	Bounce         Int64Array
	Browser        ByteArray
	BrowserVersion ByteArray
	City           ByteArray
	Country        ByteArray
	Duration       Int64Array
	EntryPage      ByteArray
	Event          ByteArray
	ExitPage       ByteArray
	Host           ByteArray
	ID             Int64Array
	Os             ByteArray
	OsVersion      ByteArray
	Path           ByteArray
	Referrer       ByteArray
	ReferrerSource ByteArray
	Region         ByteArray
	Screen         ByteArray
	Session        Int64Array
	Timestamp      Int64Array
	UtmCampaign    ByteArray
	UtmContent     ByteArray
	UtmMedium      ByteArray
	UtmSource      ByteArray
	UtmTerm        ByteArray
}

var multiPool = &sync.Pool{
	New: func() any {
		return &MultiEntry{
			Bounce:         NewInt64Array(),
			Browser:        NewByteArray(),
			BrowserVersion: NewByteArray(),
			City:           NewByteArray(),
			Country:        NewByteArray(),
			Duration:       NewInt64Array(),
			EntryPage:      NewByteArray(),
			ExitPage:       NewByteArray(),
			Host:           NewByteArray(),
			ID:             NewInt64Array(),
			Event:          NewByteArray(),
			Os:             NewByteArray(),
			OsVersion:      NewByteArray(),
			Path:           NewByteArray(),
			Referrer:       NewByteArray(),
			ReferrerSource: NewByteArray(),
			Region:         NewByteArray(),
			Screen:         NewByteArray(),
			Session:        NewInt64Array(),
			Timestamp:      NewInt64Array(),
			UtmCampaign:    NewByteArray(),
			UtmContent:     NewByteArray(),
			UtmMedium:      NewByteArray(),
			UtmSource:      NewByteArray(),
			UtmTerm:        NewByteArray(),
		}
	},
}

func NewMulti() *MultiEntry {
	return multiPool.Get().(*MultiEntry)
}

func (m *MultiEntry) Release() {
	m.Reset()
	multiPool.Put(m)
}

func (m *MultiEntry) Reset() {
	m.Bounce.Reset()
	m.Browser.Reset()
	m.BrowserVersion.Reset()
	m.City.Reset()
	m.Country.Reset()
	m.Duration.Reset()
	m.EntryPage.Reset()
	m.ExitPage.Reset()
	m.Host.Reset()
	m.ID.Reset()
	m.Event.Reset()
	m.Os.Reset()
	m.OsVersion.Reset()
	m.Path.Reset()
	m.Referrer.Reset()
	m.ReferrerSource.Reset()
	m.Region.Reset()
	m.Screen.Reset()
	m.Session.Reset()
	m.Timestamp.Reset()
	m.UtmCampaign.Reset()
	m.UtmContent.Reset()
	m.UtmMedium.Reset()
	m.UtmSource.Reset()
	m.UtmTerm.Reset()
}

func (m *MultiEntry) Append(e *Entry) {
	m.mu.Lock()
	m.Bounce.Append(e.Bounce)
	m.Browser.Append(e.Browser)
	m.BrowserVersion.Append(e.BrowserVersion)
	m.City.Append(e.City)
	m.Country.Append(e.Country)
	m.Duration.Append(int64(e.Duration))
	m.EntryPage.Append(e.EntryPage)
	m.ExitPage.Append(e.ExitPage)
	m.Host.Append(e.Host)
	m.ID.Append(int64(e.ID))
	m.Event.Append(e.Event)
	m.Os.Append(e.Os)
	m.OsVersion.Append(e.OsVersion)
	m.Path.Append(e.Path)
	m.Referrer.Append(e.Referrer)
	m.ReferrerSource.Append(e.ReferrerSource)
	m.Region.Append(e.Region)
	m.Screen.Append(e.Screen)
	m.Session.Append(e.Session)
	m.Timestamp.Append(e.Timestamp)
	m.UtmCampaign.Append(e.UtmCampaign)
	m.UtmContent.Append(e.UtmContent)
	m.UtmMedium.Append(e.UtmMedium)
	m.UtmSource.Append(e.UtmSource)
	m.UtmTerm.Append(e.UtmTerm)
	m.mu.Unlock()
}

func (m *MultiEntry) Write(f *file.Writer, r *roaring64.Bitmap) {
	g := f.AppendRowGroup()
	next := func() file.ColumnChunkWriter {
		return must.Must(g.NextColumn())("failed getting next column")
	}
	m.Bounce.Write(next())
	m.Browser.Write(next(), r, v1.Block_Index_Browser)
	m.BrowserVersion.Write(next(), r, v1.Block_Index_BrowserVersion)
	m.City.Write(next(), r, v1.Block_Index_City)
	m.Country.Write(next(), r, v1.Block_Index_Country)
	m.Duration.Write(next())
	m.EntryPage.Write(next(), r, v1.Block_Index_EntryPage)
	m.ExitPage.Write(next(), r, v1.Block_Index_ExitPage)
	m.Host.Write(next(), r, v1.Block_Index_Host)
	m.ID.Write(next())
	m.Event.Write(next(), r, v1.Block_Index_Event)
	m.Os.Write(next(), r, v1.Block_Index_Os)
	m.OsVersion.Write(next(), r, v1.Block_Index_OsVersion)
	m.Path.Write(next(), r, v1.Block_Index_Path)
	m.Referrer.Write(next(), r, v1.Block_Index_Referrer)
	m.ReferrerSource.Write(next(), r, v1.Block_Index_ReferrerSource)
	m.Region.Write(next(), r, v1.Block_Index_Region)
	m.Screen.Write(next(), r, v1.Block_Index_Screen)
	m.Session.Write(next())
	m.Timestamp.Write(next())
	m.UtmCampaign.Write(next(), r, v1.Block_Index_UtmCampaign)
	m.UtmContent.Write(next(), r, v1.Block_Index_UtmContent)
	m.UtmMedium.Write(next(), r, v1.Block_Index_UtmMedium)
	m.UtmSource.Write(next(), r, v1.Block_Index_UtmSource)
	m.UtmTerm.Write(next(), r, v1.Block_Index_UtmTerm)
	must.One(g.Close())("failed closing row group writer")
}

// Fields for constructing arrow schema on Entry.
func Fields() []arrow.Field {
	return []arrow.Field{
		{Name: "bounce", Type: arrow.PrimitiveTypes.Int64},
		{Name: "browser", Type: arrow.BinaryTypes.String},
		{Name: "browser_version", Type: arrow.BinaryTypes.String},
		{Name: "city", Type: arrow.BinaryTypes.String},
		{Name: "country", Type: arrow.BinaryTypes.String},
		{Name: "duration", Type: arrow.PrimitiveTypes.Int64},
		{Name: "entry_page", Type: arrow.BinaryTypes.String},
		{Name: "event", Type: arrow.BinaryTypes.String},
		{Name: "exit_page", Type: arrow.BinaryTypes.String},
		{Name: "host", Type: arrow.BinaryTypes.String},
		{Name: "id", Type: arrow.PrimitiveTypes.Int64},
		{Name: "os", Type: arrow.BinaryTypes.String},
		{Name: "os_version", Type: arrow.BinaryTypes.String},
		{Name: "path", Type: arrow.BinaryTypes.String},
		{Name: "referrer", Type: arrow.BinaryTypes.String},
		{Name: "referrer_source", Type: arrow.BinaryTypes.String},
		{Name: "region", Type: arrow.BinaryTypes.String},
		{Name: "screen", Type: arrow.BinaryTypes.String},
		{Name: "session", Type: arrow.PrimitiveTypes.Int64},
		{Name: "timestamp", Type: arrow.PrimitiveTypes.Int64},
		{Name: "utm_campaign", Type: arrow.BinaryTypes.String},
		{Name: "utm_content", Type: arrow.BinaryTypes.String},
		{Name: "utm_medium", Type: arrow.BinaryTypes.String},
		{Name: "utm_source", Type: arrow.BinaryTypes.String},
		{Name: "utm_term", Type: arrow.BinaryTypes.String},
	}
}

var All = Fields()

var Index = func() (m map[string]int) {
	m = make(map[string]int)
	for i := range All {
		m[All[i].Name] = i
	}
	return
}()

var ParquetSchema = parquetSchema()

func parquetSchema() *schema.Schema {
	f := Fields()
	nodes := make(schema.FieldList, len(f))
	for i := range nodes {
		x := &f[i]
		var logicalType schema.LogicalType
		var typ parquet.Type
		switch f[i].Type.ID() {
		case arrow.STRING:
			logicalType = schema.StringLogicalType{}
			typ = parquet.Types.ByteArray

		default:
			typ = parquet.Types.Int64
			logicalType = schema.NewIntLogicalType(64, true)
		}
		nodes[i] = must.Must(
			schema.NewPrimitiveNodeLogical(x.Name,
				parquet.Repetitions.Required,
				logicalType,
				typ, -1, -1),
		)("schema.NewPrimitiveNodeLogical")
	}
	root := must.Must(
		schema.NewGroupNode(parquet.DefaultRootName,
			parquet.Repetitions.Required, nodes, -1),
	)("schema.NewGroupNode")
	return schema.NewSchema(root)
}

func NewFileWriter(w io.Writer) *file.Writer {
	return file.NewParquetWriter(w,
		ParquetSchema.Root(),
		file.WithWriterProps(
			parquet.NewWriterProperties(
				parquet.WithAllocator(Pool),
				parquet.WithCompression(compress.Codecs.Zstd),
				parquet.WithCompressionLevel(10),
			),
		),
	)
}

func NewFileReader(r parquet.ReaderAtSeeker) *file.Reader {
	return must.Must(
		file.NewParquetReader(r),
	)("failed creating new parquet file reader")
}

var IndexedColumnsNames = []string{
	"browser",
	"browser_version",
	"city",
	"country",
	"entry_page",
	"event",
	"exit_page",
	"host",
	"os",
	"os_version",
	"path",
	"referrer",
	"referrer_source",
	"region",
	"screen",
	"utm_campaign",
	"utm_content",
	"utm_medium",
	"utm_source",
	"utm_term",
}

// Maps column index to column name of all indexed columns
var IndexedColumns = func() (m map[int]string) {
	m = make(map[int]string)
	for _, n := range IndexedColumnsNames {
		m[Index[n]] = n
	}
	return
}()

var Schema = arrow.NewSchema(All, nil)

func Select(names ...string) *arrow.Schema {
	o := make([]arrow.Field, len(names))
	for i := range o {
		o[i] = Schema.Field(Index[names[i]])
	}
	return arrow.NewSchema(o, nil)
}

var Pool = memory.NewGoAllocator()

var entryPool = &sync.Pool{
	New: func() any {
		return new(Entry)
	},
}

func NewEntry() *Entry {
	return entryPool.Get().(*Entry)
}

func (e *Entry) Clone() *Entry {
	o := NewEntry()
	*o = *e
	return o
}

func (e *Entry) Release() {
	*e = Entry{}
	entryPool.Put(e)
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
	e.Duration = time.UnixMilli(e.Timestamp).Sub(time.UnixMilli(s.Timestamp))
	s.Timestamp = e.Timestamp
}

func Context(ctx ...context.Context) context.Context {
	if len(ctx) > 0 {
		return compute.WithAllocator(ctx[0], Pool)
	}
	return compute.WithAllocator(context.Background(), Pool)
}

type Reader struct {
	strings []parquet.ByteArray
	ints    []int64
	b       *array.RecordBuilder
}

func NewReader() *Reader {
	return readerPool.Get().(*Reader)
}

var readerPool = &sync.Pool{
	New: func() any {
		return &Reader{
			strings: make([]parquet.ByteArray, 1<<10),
			ints:    make([]int64, 1<<10),
			b:       array.NewRecordBuilder(Pool, Schema),
		}
	},
}

func (b *Reader) Release() {
	b.strings = b.strings[:0]
	b.ints = b.ints[:0]
	readerPool.Put(b)
}

func (b *Reader) Read(r *file.Reader, groups []int) {
	x := b.b.Schema()
	if groups == nil {
		groups = make([]int, r.NumRowGroups())
		for i := range groups {
			groups[i] = i
		}
	}
	for i := range groups {
		g := r.RowGroup(groups[i])
		n := g.NumRows()
		for f := 0; f < x.NumFields(); f++ {
			b.read(f, n, must.Must(g.Column(f))("failed getting a RowGroup column"))
		}
	}
}

func (b *Reader) read(f int, rows int64, chunk file.ColumnChunkReader) {
	switch e := b.b.Field(f).(type) {
	case *array.StringBuilder:
		r := chunk.(*file.ByteArrayColumnChunkReader)
		b.strings = slices.Grow(b.strings, int(rows))[:rows]
		r.ReadBatch(rows, b.strings, nil, nil)
		e.Reserve(int(rows))
		for i := range b.strings {
			e.UnsafeAppend(b.strings[i])
		}
	case *array.Int64Builder:
		r := chunk.(*file.Int64ColumnChunkReader)
		b.ints = slices.Grow(b.ints, int(rows))[:rows]
		r.ReadBatch(rows, b.ints, nil, nil)
		e.AppendValues(b.ints, nil)
	default:
		must.AssertFMT(false)("unsupported arrow builder type %T", e)
	}
}

func (b *Reader) Record() arrow.Record {
	return b.b.NewRecord()
}
