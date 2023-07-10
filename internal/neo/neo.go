package neo

import (
	"context"
	"errors"
	"io"
	"path"
	"regexp"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/apache/arrow/go/v13/arrow/memory"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/spec"
	"golang.org/x/exp/slices"
)

var schema = parquet.SchemaOf(entry.Entry{})

// Maps field name to column number
var NameToCol = func() (o map[string]int) {
	o = make(map[string]int)
	for i, f := range schema.Fields() {
		o[f.Name()] = i
	}
	return
}()

var FilterToCol = func() (o map[string]int) {
	o = make(map[string]int)
	for i, f := range schema.Fields() {
		if _, ok := spec.ValidSegment[f.Name()]; ok {
			o[f.Name()] = i
		}
	}
	return
}()

var FieldToArrowType = func() (o map[string]arrow.Field) {
	o = make(map[string]arrow.Field)
	for _, f := range schema.Fields() {
		switch f.Type().Kind() {
		case parquet.ByteArray:
			o[f.Name()] = arrow.Field{
				Name: f.Name(),
				Type: &arrow.StringType{},
			}
		case parquet.Int64:
			switch f.Name() {
			case "timestamp":
				o[f.Name()] = arrow.Field{
					Name: f.Name(),
					Type: &arrow.TimestampType{
						Unit: arrow.Millisecond,
					},
				}
			case "duration":
				o[f.Name()] = arrow.Field{
					Name: f.Name(),
					Type: &arrow.DurationType{},
				}
			default:
				o[f.Name()] = arrow.Field{
					Name: f.Name(),
					Type: &arrow.Int64Type{},
				}
			}
		}
	}
	return
}()

var fieldPages = func() (o []Field) {
	o = make([]Field, len(schema.Fields()))
	for i, f := range schema.Fields() {
		o[i] = Field{
			Name:  f.Name(),
			Arrow: FieldToArrowType[f.Name()],
		}
	}
	return
}()

type Plan struct {
	Select []string
	Where  []string
}

type FilterType uint

const (
	EXACT FilterType = iota
	GLOB
	RE
)

type Filter struct {
	Field    string
	Value    string
	Type     FilterType
	re       *regexp.Regexp
	selected *Field
	value    parquet.Value
}

func (f *Filter) Match(v parquet.Value) bool {
	if f.Type == EXACT {
		return parquet.Equal(v, f.value)
	}
	if f.Type == GLOB {
		ok, _ := path.Match(f.Value, v.String())
		return ok
	}
	if f.re == nil {
		f.re = regexp.MustCompile(f.Value)
	}
	return f.re.MatchString(v.String())
}

type Options struct {
	Filters    []*Filter
	Start, End int64
	Select     []string
}

type GroupProcess func(g parquet.RowGroup) error

func Exec(ctx context.Context, o Options, source func(GroupProcess) error) (arrow.Record, error) {
	selectedFields := make(map[string]bool)
	for _, v := range o.Select {
		selectedFields[v] = true
	}

	pages := clonePages()

	tsCol := NameToCol["timestamp"]

	bloom := make(map[int]parquet.Value)

	for _, x := range o.Filters {
		if selectedFields[x.Field] {
			x.selected = &pages[NameToCol[x.Field]]
		}
		if x.Type == EXACT {
			x.value = parquet.ValueOf(x.Value)
			bloom[NameToCol[x.Field]] = x.value
		}
	}

	var booleans []bool
	values := make([]parquet.Value, 0, 1<<10)

	err := source(func(g parquet.RowGroup) error {
		columns := g.ColumnChunks()
		var has bool
		for k, v := range bloom {
			has, _ = columns[k].BloomFilter().Check(v)
			if !has {
				break
			}
		}
		if !has {
			// Filtering is binary AND . If one of the filter condition is not met we
			// skip this row group.
			return nil
		}
		for i := range columns {
			pages[i].Pages = columns[i].Pages()
		}
		defer func() {
			for i := range pages {
				pages[i].Pages.Close()
			}
		}()
		for {
			booleans = booleans[:0]
			values = values[:0]
			pts, err := pages[tsCol].Pages.ReadPage()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return io.EOF
				}
				return err
			}
			min, max, ok := pts.Bounds()
			if !ok {
				continue
			}
			if !bounds(o.Start, o.End, min.Int64(), max.Int64()) {
				continue
			}
			valuesInPage := pts.NumValues()
			tsValues := make([]int64, pts.NumValues())
			_, err = pts.Values().(parquet.Int64Reader).ReadInt64s(tsValues)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					return err
				}
			}
			lo, hi := filterTimestamp(tsValues, o.Start, o.End)
			booleans = slices.Grow(booleans, int(valuesInPage))[:valuesInPage]

			observedEndTs := hi < int(valuesInPage)-1

			// by default select every row from this page
			copyBool(booleans[lo:hi])

			values = slices.Grow(values, int(valuesInPage))
			match := true
			for _, x := range o.Filters {
				match = filterValues(pages[NameToCol[x.Field]].Pages, values, x, booleans)
				if !match {
					break
				}
			}
			if !match {
				if observedEndTs {
					// Nothing matched and we have seen the end timestamp. Stop looking any
					// further
					return io.EOF
				}
				continue
			}

			// select
			for i := range pages {
				x := &pages[i]
				if !selectedFields[x.Name] {
					continue
				}
				// x was selected
				x.Read(ctx, booleans)
			}
			if observedEndTs {
				return io.EOF
			}
		}
	})
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	fields := make([]arrow.Field, 0, len(selectedFields))
	var arrays []arrow.Array
	for i := range pages {
		x := pages[i]
		if selectedFields[x.Name] {
			fields = append(fields, x.Arrow)
			j, err := array.Concatenate(x.Table, memory.DefaultAllocator)
			if err != nil {
				log.Get().Fatal().Err(err).Msg("failed to join pages arrays")
			}
			arrays = append(arrays, j)
		}
	}
	schema := arrow.NewSchema(fields, nil)
	result := array.NewRecord(schema, arrays, int64(arrays[0].Len()))
	return result, nil
}

func filterTimestamp(ts []int64, start, end int64) (from, to int) {
	if ts[0] < start {
		for from = 0; from < len(ts); from++ {
			if ts[from] > start {
				break
			}
		}
	}
	if ts[len(ts)-1] > end {
		for to = len(ts) - 1; to > 0; to-- {
			if ts[to] <= end {
				break
			}
		}
	}
	return
}

func filterValues(pages parquet.Pages, values []parquet.Value, f *Filter, ok []bool) (seen bool) {
	p, err := pages.ReadPage()
	if err != nil {
		log.Get().Fatal().Err(err).Msg("corrupt data: found pages with different sizes")
	}
	_, err = p.Values().ReadValues(values)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Get().Fatal().Err(err).Msg("corrupt data: found page values with different sizes")
		}
	}
	var s *array.StringBuilder
	if f.selected != nil {
		s = f.selected.Builder.(*array.StringBuilder)
		s.Reserve(len(values))
	}
	for i := range values {
		if s != nil {
			s.Append(values[i].String())
		}
		// filtering is binary AND all filters must select the same row for it to be
		// considered.
		ok[i] = ok[i] && f.Match(values[i])
		if ok[i] {
			seen = true
		}
	}
	if s != nil {
		f.selected.Page = s.NewArray()
	}
	return
}

func bounds(start, end int64, min, max int64) bool {
	if min > end {
		return false
	}
	if start > max {
		return false
	}
	return true
}

func clonePages() FieldList {
	pages := slices.Clone(fieldPages)
	for i := range pages {
		pages[i].Builder = array.NewBuilder(memory.DefaultAllocator,
			pages[i].Arrow.Type,
		)
	}
	return pages
}

type Field struct {
	Name    string
	Pages   parquet.Pages
	Arrow   arrow.Field
	Builder array.Builder
	Page    arrow.Array
	Table   []arrow.Array
}

type FieldList []Field

func (f FieldList) Close() {
	for i := range f {
		f[i].Pages.Close()
	}
}

func (f *Field) Read(ctx context.Context, ok []bool) {
	if f.Page == nil {
		p, err := f.Pages.ReadPage()
		if err != nil {
			log.Get().Fatal().Err(err).Msg("corrupt data: found pages with different sizes")
		}
		id := f.Arrow.Type.ID()
		switch id {
		case arrow.INT64,
			arrow.DURATION,
			arrow.TIMESTAMP:
			o := make([]int64, p.NumValues())
			_, err = p.Values().(parquet.Int64Reader).ReadInt64s(o)
			if !errors.Is(err, io.EOF) {
				log.Get().Fatal().Err(err).Msg("corrupt data: found page values with different sizes")
			}
			if id == arrow.INT64 {
				f.Builder.(*array.Int64Builder).AppendValues(o, ok)
			} else if id == arrow.DURATION {
				du := make([]arrow.Duration, len(o))
				for i := range o {
					du[i] = arrow.Duration(o[i])
				}
				f.Builder.(*array.DurationBuilder).AppendValues(du, ok)
			} else if id == arrow.TIMESTAMP {
				ts := make([]arrow.Timestamp, len(o))
				for i := range o {
					ts[i] = arrow.Timestamp(o[i])
				}
				f.Builder.(*array.TimestampBuilder).AppendValues(ts, ok)
			}
		case arrow.STRING:
			o := make([]parquet.Value, p.NumValues())
			_, err = p.Values().ReadValues(o)
			if !errors.Is(err, io.EOF) {
				log.Get().Fatal().Err(err).Msg("corrupt data: found page values with different sizes")
			}
			ts := make([]string, len(o))
			for i := range o {
				ts[i] = o[i].String()
			}
			f.Builder.(*array.StringBuilder).AppendValues(ts, ok)
		default:
			log.Get().Fatal().Str("kind", f.Arrow.Type.ID().String()).
				Msg("unexpected arrow datatype")
		}
		f.Table = append(f.Table, f.Builder.NewArray())
	} else {
		a := f.Page
		sv := a.(*array.String)
		o := make([]string, a.Len())
		for i := range o {
			o[i] = sv.Value(i)
		}
		a.Release()
		f.Page = nil
		f.Builder.(*array.StringBuilder).AppendValues(o, ok)
		f.Table = append(f.Table, f.Builder.NewArray())
	}
}

var boolBuffer = func() (o []bool) {
	o = make([]bool, 5<<10)
	for i := range o {
		o[i] = true
	}
	return
}()

func copyBool(o []bool) {
	if len(o) < len(boolBuffer) {
		copy(o, boolBuffer[:len(o)])
		return
	}
	for i := range o {
		o[i] = true
	}
}
