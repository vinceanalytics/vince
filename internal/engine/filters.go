package engine

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/compute"
	"github.com/apache/arrow/go/v14/arrow/scalar"
	"github.com/bits-and-blooms/bitset"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/bloom"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/entry"
)

type Op uint

const (
	Eq Op = 1 + iota
	Ne
	Gt
	GtEg
	Lt
	LtEq
	ReEq
	ReNe
)

func (o Op) String() string {
	switch o {
	case Eq:
		return "equal"
	case Ne:
		return "not_equal"
	case Gt:
		return "greater"
	case GtEg:
		return "greater_equal"
	case Lt:
		return "less"
	case LtEq:
		return "less_equal"
	case ReEq:
		return "regex_equal"
	case ReNe:
		return "regex_not_equal"
	default:
		panic(fmt.Sprintf("unknown operation %d", o))
	}
}

var ErrNoFilter = errors.New("no filters")
var ErrSkipBlock = errors.New("skip block")

type Filters struct {
	Index []IndexFilter
	Value []ValueFilter
}

// IndexFilter is an interface for choosing row groups and row group pages to
// read based on a column index.
//
// This filter is first applied before values are read into arrow.Record. Only
// row groups and row group pages that matches across all IndexFilter are
// selected for reads
type IndexFilter interface {
	Column() v1.Column
	FilterIndex(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error)
}

type RowGroups struct {
	groups bitset.BitSet
	pages  map[uint]*bitset.BitSet
}

func NewRowGroups() *RowGroups {
	return &RowGroups{pages: make(map[uint]*bitset.BitSet)}
}

func (g *RowGroups) Empty() bool {
	return g.groups.Count() == 0
}

func (g *RowGroups) Set(index uint, pages []uint) {
	g.groups.Set(index)
	p := new(bitset.BitSet)
	for i := range pages {
		p.Set(pages[i])
	}
	g.pages[index] = p
}

func (g *RowGroups) SelectGroup(groups *bitset.BitSet, f func(uint, *bitset.BitSet)) {
	for k, v := range g.pages {
		if groups.Test(k) {
			f(k, v)
		}
	}
}

type FilterIndex func(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error)

type ValueFilter interface {
	Column() v1.Column
	FilterValue(ctx context.Context, value arrow.Array) (arrow.Array, error)
}

type FilterValue func(ctx context.Context, b *array.BooleanBuilder, value arrow.Array) (arrow.Array, error)

type IndexFilterResult struct {
	RowGroups []uint
	Pages     []*bitset.BitSet
}

func buildIndexFilter(
	ctx *sql.Context,
	f []IndexFilter, source func(*sql.Context, v1.Column) *blocksv1.ColumnIndex) (*IndexFilterResult, error) {
	if len(f) == 0 {
		return nil, ErrNoFilter
	}
	// Make sure timestamp is processed first
	sort.SliceStable(f, func(i, j int) bool {
		return f[i].Column() == v1.Column_timestamp
	})
	groups := make([]*RowGroups, len(f))
	for i := range f {
		g, err := f[i].FilterIndex(ctx, source(ctx, f[i].Column()))
		if err != nil {
			return nil, err
		}
		if g.Empty() {
			return nil, ErrSkipBlock
		}
		groups[i] = g
	}
	g := &groups[0].groups
	for i := range groups {
		if i == 0 {
			continue
		}
		g = groups[i].groups.Intersection(g)
	}
	o := make([]uint, g.Count())
	_, all := g.NextSetMany(0, o)
	slices.Sort(o)
	r := &IndexFilterResult{
		RowGroups: make([]uint, len(all)),
		Pages:     make([]*bitset.BitSet, len(all)),
	}
	pages := make(map[uint]*bitset.BitSet)
	for _, a := range all {
		pages[a] = new(bitset.BitSet)
	}
	for i := range groups {
		groups[i].SelectGroup(g, func(u uint, bs *bitset.BitSet) {
			if i == 0 {
				pages[u] = bs
			} else {
				pages[u] = pages[u].Intersection(bs)
			}
		})
	}
	for i := range all {
		r.RowGroups[i] = all[i]
		r.Pages[i] = pages[all[i]]
	}
	return r, nil
}

func applyValueFilter(ctx context.Context,
	f []ValueFilter,
	record arrow.Record,
) (arrow.Record, error) {
	if len(f) == 0 {
		return record, nil
	}
	defer record.Release()

	fields := make(map[string]arrow.Array)
	for i := 0; i < int(record.NumCols()); i++ {
		fields[record.ColumnName(i)] = record.Column(i)
	}
	source := func(c v1.Column) arrow.Array {
		return fields[c.String()]
	}
	a, err := BuildValueFilter(ctx, f, source)
	if err != nil {
		return nil, err
	}
	return compute.FilterRecordBatch(ctx, record, a, compute.DefaultFilterOptions())
}

func BuildValueFilter(ctx context.Context,
	f []ValueFilter,
	source func(v1.Column) arrow.Array) (arrow.Array, error) {
	// Make sure timestamp is processed first
	sort.SliceStable(f, func(i, j int) bool {
		return f[i].Column() == v1.Column_timestamp
	})
	var filter arrow.Array
	for i := range f {
		g, err := f[i].FilterValue(ctx, source(f[i].Column()))
		if err != nil {
			return nil, err
		}
		if filter == nil {
			filter = g
		} else {
			// we and all filters
			filter, err = entry.Call(ctx, "and", nil, filter, g, filter.Release, g.Release)
			if err != nil {
				return nil, err
			}
		}
	}

	return filter, nil
}

type FilterIndexFunc func(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error)

type IndexMatchFuncs struct {
	Col             v1.Column
	Value           any
	FilterIndexFunc FilterIndexFunc
}

var _ IndexFilter = (*IndexMatchFuncs)(nil)

func (i *IndexMatchFuncs) Column() v1.Column {
	return i.Col
}

func (i *IndexMatchFuncs) FilterIndex(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error) {
	return i.FilterIndexFunc(ctx, idx)
}

func buildValueFilterMatch(col v1.Column, lo, hi any, op Op) *ValueMatchFuncs {
	var value any
	switch op {
	case Eq, Gt, GtEg:
		value = lo
	case Lt, LtEq:
		value = hi
	}
	if value == nil {
		return nil
	}
	return Match(col, value, op)
}

func buildIndexFilterMatch(col v1.Column, lo, hi any, op Op) *IndexMatchFuncs {
	var value any
	switch op {
	case Eq, Gt, GtEg:
		value = lo
	case Lt, LtEq:
		value = hi
	}
	if value == nil || col == -1 {
		return nil
	}
	return &IndexMatchFuncs{
		Col:   col,
		Value: value,
		FilterIndexFunc: func(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error) {
			if col == v1.Column_timestamp {
				return filterTimestamp(value.(time.Time), op)(ctx, idx)
			}
			return filterBloom(value)(ctx, idx)
		},
	}
}

func filterTimestamp(timestamp time.Time, op Op) FilterIndexFunc {
	return func(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error) {
		ts := timestamp.UTC().UnixMilli()
		if !timeInRange(ts, idx.Min, idx.Max, op) {
			return nil, ErrSkipBlock
		}
		g := NewRowGroups()
		pages := make([]uint, 0, 32)
		for i := len(idx.RowGroups) - 1; i >= 0; i-- {
			rg := idx.RowGroups[i]
			if !timeInRange(ts, rg.Min, rg.Max, op) {
				break
			}

			pages = slices.Grow(pages, len(rg.Pages))[:0]
			for j := len(rg.Pages) - 1; j >= 0; j-- {
				page := rg.Pages[j]
				if !timeInRange(ts, page.Min, page.Max, op) {
					break
				}
				pages = append(pages, uint(j))
			}
			g.Set(uint(i), pages)
		}
		return g, nil
	}
}

func filterBloom(value any) FilterIndexFunc {
	h := bloom.XXH64{}
	v := parquet.ValueOf(value)
	var hash uint64
	switch v.Kind() {
	case parquet.Int64, parquet.Double:
		hash = h.Sum64Uint64(v.Uint64())
	default:
		hash = h.Sum64(v.ByteArray())
	}
	return func(ctx context.Context, idx *blocksv1.ColumnIndex) (*RowGroups, error) {
		g := NewRowGroups()
		pages := make([]uint, 0, 1<<10)
		for i := len(idx.RowGroups) - 1; i >= 0; i-- {
			rg := idx.RowGroups[i]
			if !inBloom(hash, rg.BloomFilter) {
				continue
			}
			pages = slices.Grow(pages, len(rg.Pages))[:0]
			for j := range rg.Pages {
				pages = append(pages, uint(j))
			}
			g.Set(uint(i), pages)
		}
		return g, nil
	}
}

func inBloom(hash uint64, b []byte) bool {
	return bloom.MakeSplitBlockFilter(b).Check(hash)
}

func timeInRange(timestamp int64, a, b int64, op Op) bool {
	switch op {
	case Lt:
		return a <= timestamp
	case LtEq:
		return a < timestamp
	case Gt:
		return b >= timestamp
	case GtEg:
		return b > timestamp
	default:
		return a < timestamp && timestamp < b
	}
}

type ValueMatchFuncs struct {
	Col             v1.Column
	Value           any
	FilterValueFunc func(ctx context.Context, value arrow.Array) (arrow.Array, error)
}

var _ ValueFilter = (*ValueMatchFuncs)(nil)

func (i *ValueMatchFuncs) Column() v1.Column {
	return i.Col
}

func (i *ValueMatchFuncs) FilterValue(ctx context.Context, value arrow.Array) (arrow.Array, error) {
	return i.FilterValueFunc(ctx, value)
}

func Match(col v1.Column, matchValue any, op Op) *ValueMatchFuncs {
	return &ValueMatchFuncs{
		Col: col,
		FilterValueFunc: func(ctx context.Context, value arrow.Array) (arrow.Array, error) {
			var fv scalar.Scalar
			if ts, ok := matchValue.(time.Time); ok {
				fv = scalar.NewTimestampScalar(
					arrow.Timestamp(ts.UnixMilli()),
					arrow.FixedWidthTypes.Timestamp_ms,
				)
			} else {
				fv = scalar.MakeScalar(matchValue)
			}
			return entry.Call(ctx, op.String(), nil, value, &compute.ScalarDatum{
				Value: fv,
			})
		},
	}
}
