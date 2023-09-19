package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/parquet-go/parquet-go"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	vdb "github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"golang.org/x/sync/errgroup"
)

var Columns, Indexed = func() (o []storev1.Column, idx map[string]storev1.Column) {
	idx = make(map[string]storev1.Column)
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		idx[i.String()] = i
		o = append(o, i)
	}
	return
}()

type tableSchema struct {
	sql   sql.Schema
	arrow *arrow.Schema
}

func (ts *tableSchema) read(ctx context.Context, r *parquet.File) (arrow.Record, error) {
	b := array.NewRecordBuilder(entry.Pool, ts.arrow)
	defer b.Release()

	fields := ts.arrow.Fields()
	schema := r.Schema()
	fieldToColIdx := make(map[string]int)
	for i := range fields {
		column, ok := schema.Lookup(fields[i].Name)
		if !ok {
			return nil, fmt.Errorf("column %q not found in parquet file", fields[i].Name)
		}
		fieldToColIdx[fields[i].Name] = column.ColumnIndex
	}

	var eg errgroup.Group

	for _, g := range r.RowGroups() {
		chunks := g.ColumnChunks()
		for i := range fields {
			eg.Go(ts.readColum(
				ctx,
				&fields[i], b.Field(i),
				chunks[fieldToColIdx[fields[i].Name]],
			))
		}
	}
	err := eg.Wait()
	if err != nil {
		return nil, err
	}
	return b.NewRecord(), nil
}

func (ts *tableSchema) readColum(
	ctx context.Context,
	field *arrow.Field,
	b array.Builder,
	chunk parquet.ColumnChunk,
) func() error {
	return func() error {
		buf, err := readValuesPages(chunk.Pages())
		if err != nil {
			return err
		}
		defer buf.release()
		b.Reserve(len(buf.values))
		switch e := b.(type) {
		case *array.Int64Builder:
			for i := range buf.values {
				e.UnsafeAppend(buf.values[i].Int64())
			}
		case *array.TimestampBuilder:
			for i := range buf.values {
				e.UnsafeAppend(arrow.Timestamp(buf.values[i].Int64()))
			}
		case *array.Float64Builder:
			for i := range buf.values {
				e.UnsafeAppend(time.Duration(buf.values[i].Int64()).Seconds())
			}
		case *array.StringBuilder:
			for i := range buf.values {
				e.Append(buf.values[i].String())
			}
		}
		return nil
	}
}

func readValuesPages(pages parquet.Pages) (*valuesBuf, error) {
	defer pages.Close()
	buf := valuesBufPool.Get().(*valuesBuf)
	for {
		page, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return buf, nil
			}
			buf.release()
			return nil, err
		}
		size := page.NumValues()
		o := buf.get(int(size))
		page.Values().ReadValues(o)
	}
}

type valuesBuf struct {
	values []parquet.Value
}

func (i *valuesBuf) release() {
	i.values = i.values[:0]
	valuesBufPool.Put(i)
}

func (i *valuesBuf) get(n int) []parquet.Value {
	x := len(i.values)
	i.values = slices.Grow(i.values, n)
	i.values = i.values[:x+n]
	return i.values[x:n]
}

var valuesBufPool = &sync.Pool{New: func() any { return &valuesBuf{} }}

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

func createSchema(table string, columns []storev1.Column) (o tableSchema) {
	fields := make([]arrow.Field, 0, len(columns))
	for _, i := range columns {
		if i <= storev1.Column_timestamp {
			switch i {
			case storev1.Column_duration:
				o.sql = append(o.sql, &sql.Column{
					Name:   i.String(),
					Type:   types.Float64,
					Source: table,
				})
				fields = append(fields, arrow.Field{
					Name: i.String(),
					Type: arrow.PrimitiveTypes.Float64,
				})
			case storev1.Column_timestamp:
				o.sql = append(o.sql, &sql.Column{
					Name:   i.String(),
					Type:   types.Timestamp,
					Source: table,
				})
				fields = append(fields, arrow.Field{
					Name: i.String(),
					Type: arrow.FixedWidthTypes.Timestamp_ms,
				})
			default:
				o.sql = append(o.sql, &sql.Column{
					Name:   i.String(),
					Type:   types.Int64,
					Source: table,
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
			Source:   table,
		})
		fields = append(fields, arrow.Field{
			Name: i.String(),
			Type: arrow.BinaryTypes.String,
		})
	}
	o.arrow = arrow.NewSchema(fields, nil)
	return
}

type DB struct {
	Context
}

var _ sql.Database = (*DB)(nil)

func (DB) Name() string {
	return "vince"
}

func (db *DB) GetTableInsensitive(ctx *sql.Context, tblName string) (table sql.Table, ok bool, err error) {
	db.DB.Txn(false, func(txn vdb.Txn) error {
		key := keys.Site(tblName)
		if txn.Has(key) {
			table = &Table{Context: db.Context,
				name:   tblName,
				schema: createSchema(tblName, Columns)}
			ok = true
		}
		return nil
	})
	return
}

func (db *DB) GetTableNames(ctx *sql.Context) (names []string, err error) {
	db.DB.Txn(false, func(txn vdb.Txn) error {
		key := keys.Site("")
		it := txn.Iter(vdb.IterOpts{
			Prefix: key,
		})
		for it.Rewind(); it.Valid(); it.Next() {
			names = append(names,
				string(bytes.TrimPrefix(it.Key(), key)))
		}
		return nil
	})
	return
}

func (DB) IsReadOnly() bool {
	return true
}

var _ sql.DatabaseProvider = (*Provider)(nil)
var _ sql.FunctionProvider = (*Provider)(nil)

type Provider struct {
	Context
}

func (p *Provider) Function(ctx *sql.Context, name string) (sql.Function, error) {
	fn, ok := funcs[strings.ToLower(name)]
	if !ok {
		return nil, sql.ErrFunctionNotFound.New(name)
	}
	return fn, nil
}

func (p *Provider) Database(_ *sql.Context, name string) (sql.Database, error) {
	if name != "vince" {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return &DB{Context: p.Context}, nil
}

func (p *Provider) AllDatabases(_ *sql.Context) []sql.Database {
	return []sql.Database{&DB{Context: p.Context}}
}

func (p *Provider) HasDatabase(_ *sql.Context, name string) bool {
	return name == "vince"
}
