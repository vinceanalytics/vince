package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/bits-and-blooms/bitset"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/parquet-go/parquet-go"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/entry"
	"golang.org/x/sync/errgroup"
)

var Columns, Indexed = func() (o []string, idx map[string]storev1.Column) {
	idx = make(map[string]storev1.Column)
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		if i >= storev1.Column_timestamp {
			idx[i.String()] = i
		}
		o = append(o, i.String())
	}
	return
}()

type tableSchema struct {
	sql   sql.Schema
	arrow *arrow.Schema
}

func (ts *tableSchema) read(ctx context.Context,
	domain string,
	r *parquet.File,
	rowGroups []uint,
	pages []*bitset.BitSet,
) (arrow.Record, error) {
	b := array.NewRecordBuilder(entry.Pool, ts.arrow)
	defer b.Release()
	fields := ts.arrow.Fields()
	schema := r.Schema()
	fieldToColIdx := make(map[string]int)
	name := -1
	for i := range fields {
		if fields[i].Name == "name" {
			name = i
			continue
		}
		column, ok := schema.Lookup(fields[i].Name)
		if !ok {
			return nil, fmt.Errorf("column %q not found in parquet file", fields[i].Name)
		}
		fieldToColIdx[fields[i].Name] = column.ColumnIndex
	}
	groups := r.RowGroups()

	var eg errgroup.Group

	accept := make(map[uint]int)
	for i, g := range rowGroups {
		accept[g] = i
	}

	for i, g := range groups {
		n, ok := accept[uint(i)]
		if !ok {
			continue
		}
		page := pages[n]
		chunks := g.ColumnChunks()
		for i := range fields {
			if fields[i].Name == "name" {
				continue
			}
			eg.Go(ts.readColum(
				ctx,
				&fields[i], b.Field(i),
				chunks[fieldToColIdx[fields[i].Name]],
				page.Clone(),
			))
		}
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	if name != -1 {
		// domain name is not stored as part of the parquet file. We need to manually
		// add the name column when it is selected.
		for i := range fields {
			if i != name {
				nb := b.Field(name).(*array.StringBuilder)
				size := b.Field(i).Len()
				nb.Reserve(size)
				for n := 0; n < size; n++ {
					nb.Append(domain)
				}
				break
			}
		}
	}
	return b.NewRecord(), nil
}

func (ts *tableSchema) readColum(
	ctx context.Context,
	field *arrow.Field,
	b array.Builder,
	chunk parquet.ColumnChunk,
	accept *bitset.BitSet,
) func() error {
	return func() error {
		buf, err := readValuesPages(chunk.Pages(), accept)
		if err != nil {
			return err
		}
		defer buf.Release()
		b.Reserve(len(buf.Values))
		switch e := b.(type) {
		case *array.Int64Builder:
			for i := range buf.Values {
				e.UnsafeAppend(buf.Values[i].Int64())
			}
		case *array.TimestampBuilder:
			for i := range buf.Values {
				e.UnsafeAppend(arrow.Timestamp(buf.Values[i].Int64()))
			}
		case *array.Float64Builder:
			for i := range buf.Values {
				e.UnsafeAppend(time.Duration(buf.Values[i].Int64()).Seconds())
			}
		case *array.StringBuilder:
			for i := range buf.Values {
				e.Append(buf.Values[i].String())
			}
		}
		return nil
	}
}

func readValuesPages(pages parquet.Pages, accept *bitset.BitSet) (*entry.ValuesBuf, error) {
	defer pages.Close()
	buf := entry.NewValuesBuf()
	for i := uint(0); ; i++ {
		page, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return buf, nil
			}
			buf.Release()
			return nil, err
		}
		if !accept.Test(i) {
			continue
		}
		size := page.NumValues()
		o := buf.Get(int(size))
		page.Values().ReadValues(o)
	}
}

type recordIter struct {
	pos        int
	rows, cols int
	values     sql.Row
	record     arrow.Record
}

var _ sql.RowIter = (*recordIter)(nil)

func (r *recordIter) Next(ctx *sql.Context) (sql.Row, error) {
	r.pos++
	if r.pos < r.rows {
		for i := range r.values {
			r.values[i] = r.value(
				r.record.Column(i),
				r.pos,
			)
		}
		return r.values, nil
	}
	return nil, io.EOF
}

func (r *recordIter) value(a arrow.Array, idx int) any {
	switch e := a.(type) {
	case *array.Int64:
		return e.Value(idx)
	case *array.Timestamp:
		return e.Value(idx).ToTime(arrow.Millisecond)
	case *array.Float64:
		return e.Value(idx)
	case *array.String:
		return e.Value(idx)
	default:
		panic(fmt.Sprintf("unsupported array data type %#T", e))
	}
}

func (r *recordIter) Close(*sql.Context) error {
	r.record.Release()
	r.record = nil
	r.pos, r.cols, r.rows = 0, 0, 0
	r.values = r.values[:0]
	recordIterPool.Put(r)
	return nil
}

var recordIterPool = &sync.Pool{New: func() any { return new(recordIter) }}

func newRecordIter(r arrow.Record) *recordIter {
	x := recordIterPool.Get().(*recordIter)
	x.pos = -1
	x.rows = int(r.NumRows())
	x.cols = int(r.NumCols())
	x.values = slices.Grow(x.values, x.cols)[:x.cols]
	x.record = r
	return x
}

const eventsTableName = "events"

func createSchema(columns []string) (o tableSchema) {
	// Make timestamp the first column we read. This ensures we pick the right
	// pages to read from the blocks
	ts := storev1.Column_timestamp.String()
	sort.SliceStable(columns, func(i, j int) bool {
		return columns[i] == ts
	})
	fields := make([]arrow.Field, 0, len(columns))
	for _, col := range columns {
		i := v1.Column(v1.Column_value[col])
		if i <= storev1.Column_timestamp {
			switch i {
			case storev1.Column_duration:
				o.sql = append(o.sql, &sql.Column{
					Name:   i.String(),
					Type:   types.Float64,
					Source: eventsTableName,
				})
				fields = append(fields, arrow.Field{
					Name: i.String(),
					Type: arrow.PrimitiveTypes.Float64,
				})
			case storev1.Column_timestamp:
				o.sql = append(o.sql, &sql.Column{
					Name:   i.String(),
					Type:   types.Timestamp,
					Source: eventsTableName,
				})
				fields = append(fields, arrow.Field{
					Name: i.String(),
					Type: arrow.FixedWidthTypes.Timestamp_ms,
				})
			default:
				o.sql = append(o.sql, &sql.Column{
					Name:   i.String(),
					Type:   types.Int64,
					Source: eventsTableName,
				})
				fields = append(fields, arrow.Field{
					Name: i.String(),
					Type: arrow.PrimitiveTypes.Int64,
				})
			}
			continue
		}
		o.sql = append(o.sql, &sql.Column{
			Name:     i.String(),
			Type:     types.Text,
			Nullable: false,
			Source:   eventsTableName,
		})
		fields = append(fields, arrow.Field{
			Name: i.String(),
			Type: arrow.BinaryTypes.String,
		})
	}
	o.arrow = arrow.NewSchema(fields, nil)
	return
}
