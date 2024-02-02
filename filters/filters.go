package filters

import (
	"github.com/blevesearch/vellum/regexp"
	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
)

const (
	TimeUnixNano       = "TimeUnixNano"
	ResourceSchema     = "ResourceSchema"
	ResourceAttributes = "ResourceAttributes"
	ScopeName          = "ScopeName"
	ScopeSchema        = "ScopeSchema"
	ScopeVersion       = "ScopeVersion"
	ScopeAttributes    = "ScopeAttributes"
	ScopeHash          = "ScopeHash"
	Name               = "Name"
	AttributesColumn   = "Attributes"
	TraceID            = "TraceID"
)

type CompiledFilter struct {
	Base  *v1.Filter
	Value []byte
	Re    *regexp.Regexp
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
	o := &CompiledFilter{Base: f}
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
