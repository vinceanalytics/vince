package filters

import (
	"github.com/blevesearch/vellum/regexp"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
)

func Column(p v1.Property) string {
	return p.String()
}

func ToColumn(p v1.Filters_Projection) string {
	return p.String()
}

func Projection(p v1.Property) v1.Filters_Projection {
	return v1.Filters_Projection(v1.Filters_Projection_value[p.String()])
}

type CompiledFilter struct {
	Column string
	Op     v1.Filter_OP
	Value  []byte
	Re     *regexp.Regexp
}

func CompileFilters(f *v1.Filters) ([]*CompiledFilter, error) {
	o := make([]*CompiledFilter, 0, len(f.List))
	for _, v := range f.List {
		x, err := compileFilter(v)
		if err != nil {
			return nil, err
		}
		o = append(o, x)
	}
	return o, nil
}

func compileFilter(f *v1.Filter) (*CompiledFilter, error) {
	o := &CompiledFilter{
		Column: Column(f.Property),
		Op:     f.Op,
	}
	o.Value = []byte(f.Value)
	switch f.Op {
	case v1.Filter_re_equal, v1.Filter_re_not_equal:
		re, err := regexp.New(f.Value)
		if err != nil {
			return nil, err
		}
		o.Re = re
	}
	return o, nil
}
