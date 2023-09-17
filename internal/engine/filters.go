package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/compute"
	"github.com/apache/arrow/go/v14/arrow/scalar"
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

type Filter interface {
	Column() v1.Column
}

type IndexFilter interface {
	Filter
	// Returns row groups to read from in the index
	FilterIndex(ctx context.Context, rowGroups []int) ([]int, error)
}

type FilterIndex func(ctx context.Context, rowGroups []int) ([]int, error)

type ValueFilter interface {
	Filter
	FilterValue(ctx context.Context, value arrow.Array) (arrow.Array, error)
}

type FilterValue func(ctx context.Context, b *array.BooleanBuilder, value arrow.Array) (arrow.Array, error)

func BuildIndexFilter(
	ctx context.Context,
	f []IndexFilter, source func(v1.Column) []int) ([]int, error) {
	// Make sure timestamp is processed first
	sort.Slice(f, func(i, j int) bool {
		return f[i].Column() == v1.Column_timestamp
	})
	seen := make(map[int]struct{})
	for i := range f {
		g, err := f[i].FilterIndex(ctx, source(f[i].Column()))
		if err != nil {
			return nil, err
		}
		if i == 0 {
			// FIrst iteration select all row groups returned by this filter
			for _, n := range g {
				seen[n] = struct{}{}
			}
		} else {
			for _, n := range g {
				_, ok := seen[n]
				if !ok {
					delete(seen, n)
				}
			}
		}
	}
	o := make([]int, 0, len(seen))
	for k := range seen {
		o = append(o, k)
	}
	sort.Ints(o)
	return o, nil
}

func BuildValueFilter(ctx context.Context,
	f []ValueFilter,
	source func(v1.Column) arrow.Array) (arrow.Array, error) {
	// Make sure timestamp is processed first
	sort.Slice(f, func(i, j int) bool {
		return f[i].Column() == v1.Column_timestamp
	})
	var filter arrow.Array
	for i := range f {
		g, err := f[i].FilterValue(ctx, source(f[i].Column()))
		if err != nil {
			return nil, err
		}
		if i == 0 {
			filter = g
		} else {
			// we and all filters
			filter, err = call("and", nil, filter, g, filter.Release, g.Release)
			if err != nil {
				return nil, err
			}
		}
	}

	return filter, nil
}

func call(name string, o compute.FunctionOptions, a any, b any, fn ...func()) (arrow.Array, error) {
	ad := compute.NewDatum(a)
	bd := compute.NewDatum(b)
	defer ad.Release()
	defer bd.Release()
	defer func() {
		for _, f := range fn {
			f()
		}
	}()
	out, err := compute.CallFunction(context.TODO(), name, o, ad, bd)
	if err != nil {
		return nil, err
	}
	return out.(*compute.ArrayDatum).MakeArray(), nil
}

type IndexMatchFuncs struct {
	Col             v1.Column
	FilterIndexFunc func(ctx context.Context, rowGroups []int) ([]int, error)
}

var _ IndexFilter = (*IndexMatchFuncs)(nil)

func (i *IndexMatchFuncs) Column() v1.Column {
	return i.Col
}

func (i *IndexMatchFuncs) FilterIndex(ctx context.Context, rowGroups []int) ([]int, error) {
	return i.FilterIndexFunc(ctx, rowGroups)
}

type ValueMatchFuncs struct {
	Col             v1.Column
	FilterValueFunc func(ctx context.Context, value arrow.Array) (arrow.Array, error)
}

var _ ValueFilter = (*ValueMatchFuncs)(nil)

func (i *ValueMatchFuncs) Column() v1.Column {
	return i.Col
}

func (i *ValueMatchFuncs) FilterValue(ctx context.Context, value arrow.Array) (arrow.Array, error) {
	return i.FilterValueFunc(ctx, value)
}

func Match[T int64 | arrow.Timestamp | float64 | string](col v1.Column, matchValue any, op Op) ValueFilter {
	return &ValueMatchFuncs{
		Col: col,
		FilterValueFunc: func(ctx context.Context, value arrow.Array) (arrow.Array, error) {
			switch op {
			case ReEq:
				m, ok := any(matchValue).(string)
				if !ok {
					slog.Warn("using regex for not string columns is not supported",
						"column", col.String(),
						"value", any(matchValue),
					)
					return nil, ErrNoFilter
				}
				return boolExpr(value.(*array.String), reMatch(m))
			case ReNe:
				m, ok := any(matchValue).(string)
				if !ok {
					slog.Warn("using regex for not string columns is not supported",
						"column", col.String(),
						"value", any(matchValue),
					)
					return nil, ErrNoFilter
				}
				return boolExpr(value.(*array.String), not(reMatch(m)))
			default:
				return call(op.String(), nil, value, &compute.ScalarDatum{
					Value: scalar.MakeScalar(matchValue),
				})
			}
		},
	}
}

func not(m func(string) bool) func(string) bool {
	return func(s string) bool {
		return !m(s)
	}
}

func reMatch(r string) func(s string) bool {
	x := regexp.MustCompile(r)
	return x.MatchString
}

func boolExpr(s *array.String, f func(string) bool) (arrow.Array, error) {
	b := array.NewBooleanBuilder(entry.Pool)
	defer b.Release()
	b.Reserve(s.Len())
	for i := 0; i < s.Len(); i++ {
		b.UnsafeAppend(f(s.Value(i)))
	}
	return b.NewBooleanArray(), nil
}
