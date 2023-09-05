package entry

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/compute"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/compress"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/schema"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/px"
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
	Timestamp      int64         `parquet:"timestamp,dict,zstd"`
	UtmCampaign    string        `parquet:"utm_campaign,dict,zstd"`
	UtmContent     string        `parquet:"utm_content,dict,zstd"`
	UtmMedium      string        `parquet:"utm_medium,dict,zstd"`
	UtmSource      string        `parquet:"utm_source,dict,zstd"`
	UtmTerm        string        `parquet:"utm_term,dict,zstd"`
}

type MultiEntry struct {
	ints    [storev1.Column_timestamp + 1][]int64
	strings [storev1.Column_utm_term - storev1.Column_timestamp][]parquet.ByteArray
}

var multiPool = &sync.Pool{
	New: func() any {
		return &MultiEntry{}
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
	for i := range m.ints {
		m.ints[i] = m.ints[i][:0]
	}
	for i := range m.strings {
		m.strings[i] = m.strings[i][:0]
	}
}

func (m *MultiEntry) Len() int {
	return len(m.ints[0])
}

func (m *MultiEntry) Append(e *Entry) {
	m.add(storev1.Column_bounce, e.Bounce)
	m.add(storev1.Column_browser, e.Browser)
	m.add(storev1.Column_browser_version, e.BrowserVersion)
	m.add(storev1.Column_city, e.City)
	m.add(storev1.Column_country, e.Country)
	m.add(storev1.Column_duration, int64(e.Duration))
	m.add(storev1.Column_entry_page, e.EntryPage)
	m.add(storev1.Column_event, e.Event)
	m.add(storev1.Column_exit_page, e.ExitPage)
	m.add(storev1.Column_host, e.Host)
	m.add(storev1.Column_id, int64(e.ID))
	m.add(storev1.Column_os, e.Os)
	m.add(storev1.Column_os_version, e.OsVersion)
	m.add(storev1.Column_path, e.Path)
	m.add(storev1.Column_referrer, e.Referrer)
	m.add(storev1.Column_referrer_source, e.ReferrerSource)
	m.add(storev1.Column_region, e.Region)
	m.add(storev1.Column_screen, e.Screen)
	m.add(storev1.Column_session, e.Session)
	m.add(storev1.Column_timestamp, e.Timestamp)
	m.add(storev1.Column_utm_campaign, e.UtmCampaign)
	m.add(storev1.Column_utm_content, e.UtmContent)
	m.add(storev1.Column_utm_medium, e.UtmMedium)
	m.add(storev1.Column_utm_source, e.UtmSource)
	m.add(storev1.Column_utm_term, e.UtmTerm)
}

func (m *MultiEntry) add(name storev1.Column, v any) {
	if name <= storev1.Column_timestamp {
		m.ints[name] = append(m.ints[name], v.(int64))
		return
	}
	m.strings[px.ColumnIndex(name)] =
		append(m.strings[px.ColumnIndex(name)], parquet.ByteArray(v.(string)))
}

func (m *MultiEntry) Write(f *file.Writer, hash func(storev1.Column, parquet.ByteArray)) {
	if m.Len() == 0 {
		return
	}
	g := f.AppendRowGroup()
	nextInt := func(v []int64) {
		x := must.Must(g.NextColumn())("failed getting next column")
		w := x.(*file.Int64ColumnChunkWriter)
		must.Must(w.WriteBatch(v, nil, nil))(
			"failed writing int64 column to parquet",
		)
		must.One(w.Close())("failed closing column writer")
	}
	nextString := func(c storev1.Column, v []parquet.ByteArray) {
		x := must.Must(g.NextColumn())("failed getting next column")
		w := x.(*file.ByteArrayColumnChunkWriter)
		must.Must(w.WriteBatch(v, nil, nil))(
			"failed writing int64 column to parquet",
		)
		must.One(w.Close())("failed closing column writer")
	}
	for i := range m.ints {
		nextInt(m.ints[i])
	}
	for i := range m.strings {
		nextString(storev1.Column(i+len(m.ints)), m.strings[i])
	}
	must.One(g.Close())("failed closing row group writer")
}

// Fields for constructing arrow schema on Entry.
func Fields() (f []arrow.Field) {
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		if i <= storev1.Column_timestamp {
			f = append(f, arrow.Field{
				Name: i.String(),
				Type: arrow.PrimitiveTypes.Int64,
			})
			continue
		}
		f = append(f, arrow.Field{
			Name: i.String(),
			Type: arrow.BinaryTypes.String,
		})
	}
	return
}

var All = Fields()

var ParquetSchema = parquetSchema()

func parquetSchema() *schema.Schema {
	f := Fields()
	nodes := make(schema.FieldList, 0, len(f))
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		if i <= storev1.Column_timestamp {
			nodes = append(nodes, must.Must(
				schema.NewPrimitiveNodeLogical(i.String(),
					parquet.Repetitions.Required,
					schema.NewIntLogicalType(64, true),
					parquet.Types.Int64, -1, -1),
			)("schema.NewPrimitiveNodeLogical"))
			continue
		}
		nodes = append(nodes, must.Must(
			schema.NewPrimitiveNodeLogical(i.String(),
				parquet.Repetitions.Required,
				schema.StringLogicalType{},
				parquet.Types.ByteArray, -1, -1),
		)("schema.NewPrimitiveNodeLogical"))
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
				parquet.WithStats(true),
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

var Schema = arrow.NewSchema(All, nil)

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

type colReader interface {
	Read(int64, file.ColumnChunkReader)
	ReadResult
}

type stringReader struct {
	col    storev1.Column
	values []parquet.ByteArray
}

var _ colReader = (*stringReader)(nil)

func (r *stringReader) Read(size int64, rd file.ColumnChunkReader) {
	x := rd.(*file.ByteArrayColumnChunkReader)
	v := make([]parquet.ByteArray, size)
	x.ReadBatch(size, v, nil, nil)
	r.values = append(r.values, v...)
}

func (r *stringReader) Value(i int) any {
	if i < len(r.values) {
		return r.values[i]
	}
	return nil
}

func (r *stringReader) Len() int {
	return len(r.values)
}

func (r *stringReader) Col() storev1.Column {
	return r.col
}

type int64Reader struct {
	col    storev1.Column
	values []int64
}

var _ colReader = (*stringReader)(nil)

func (r *int64Reader) Read(size int64, rd file.ColumnChunkReader) {
	x := rd.(*file.Int64ColumnChunkReader)
	v := make([]int64, size)
	x.ReadBatch(size, v, nil, nil)
	r.values = append(r.values, v...)
}

func (r *int64Reader) Value(i int) any {
	if i < len(r.values) {
		return r.values[i]
	}
	return nil
}

func (r *int64Reader) Len() int {
	return len(r.values)
}

func (r *int64Reader) Col() storev1.Column {
	return r.col
}

type ReadResult interface {
	Col() storev1.Column
	Len() int
	Value(int) any
}

func ReadColumns(r *file.Reader, columns []storev1.Column, groups []int) (o []ReadResult) {
	if len(columns) == 0 {
		columns = make([]storev1.Column, 0, storev1.Column_utm_term+1)
		for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
			columns = append(columns, i)
		}
	}
	cr := make([]colReader, 0, len(columns))
	for _, i := range columns {
		if i <= storev1.Column_timestamp {
			cr = append(cr, &int64Reader{
				col: i,
			})
			continue
		}
		cr = append(cr, &stringReader{
			col: i,
		})
	}
	for i := range groups {
		g := r.RowGroup(groups[i])
		size := g.NumRows()
		for n := range columns {
			rd := must.Must(g.Column(int(columns[n])))("failed getting column from row group")
			cr[n].Read(size, rd)
		}
	}
	o = make([]ReadResult, len(cr))
	for i := range cr {
		o[i] = cr[i]
	}
	return
}
