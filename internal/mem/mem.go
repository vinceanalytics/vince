package mem

import (
	"errors"
	"io"
	"slices"
	"time"

	"github.com/parquet-go/parquet-go"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/entry"
)

func ReadColumns(r *parquet.File, columns []storev1.Column, groups []int) (o []entry.ReadResult) {
	if len(columns) == 0 {
		columns = make([]storev1.Column, 0, storev1.Column_utm_term+1)
		for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
			columns = append(columns, i)
		}
	}
	schema := r.Schema()

	o = make([]entry.ReadResult, len(columns))
	idx := make([]int, len(columns))
	for i := range o {
		o[i] = &ColReader{
			column: columns[i],
			values: make([]parquet.Value, 0, 64),
		}
		l, _ := schema.Lookup(columns[i].String())
		idx[i] = l.ColumnIndex
	}
	rg := r.RowGroups()

	for i := range groups {
		g := rg[groups[i]]
		gc := g.ColumnChunks()
		size := g.NumRows()
		for n := range columns {
			o[n].(*ColReader).Read(size, gc[idx[n]])
		}
	}
	return
}

type ColReader struct {
	column storev1.Column
	values []parquet.Value
}

func (c *ColReader) Col() storev1.Column {
	return c.column
}

func (c *ColReader) Len() int {
	return len(c.values)
}

func (c *ColReader) Value(n int) any {
	if n >= len(c.values) {
		return nil
	}
	v := c.values[n]
	switch c.column {
	case storev1.Column_timestamp:
		return time.UnixMilli(v.Int64())
	case storev1.Column_bounce,
		storev1.Column_session, storev1.Column_id:
		return v.Int64()
	case storev1.Column_duration:
		return time.Duration(v.Int64()).Seconds()
	default:
		return v.String()
	}
}

func (c *ColReader) Read(size int64, col parquet.ColumnChunk) error {
	c.values = slices.Grow(c.values, len(c.values)+int(size))
	pages := col.Pages()
	for {
		page, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		valueSize := page.NumRows()
		pos := len(c.values)
		c.values = c.values[:pos+int(valueSize)]
		_, err = page.Values().ReadValues(c.values[pos:])
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil
			}
		}

	}
}
