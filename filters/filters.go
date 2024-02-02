package filters

import (
	"bytes"

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

type Op uint

const (
	Equal Op = iota
	NotEqual
	ReMatch
	ReNotMatch
	Latest
)

type CompiledFilter struct {
	Column string
	Key    string
	Op     Op
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
	var o CompiledFilter
	switch e := f.Column.(type) {
	case *v1.Filter_Base:
		o.Column = e.Base.String()
	case *v1.Filter_ResourceAttributes:
		o.Column = ResourceAttributes
		o.Key = e.ResourceAttributes
	case *v1.Filter_ScopeAttributes:
		o.Column = ScopeAttributes
		o.Key = e.ScopeAttributes
	case *v1.Filter_Attributes:
		o.Column = AttributesColumn
		o.Key = e.Attributes
	}
	switch e := f.Value.(type) {
	case *v1.Filter_Equal:
		o.Value = bytes.Clone(e.Equal)
		o.Op = Equal
	case *v1.Filter_NotEqual:
		o.Value = bytes.Clone(e.NotEqual)
		o.Op = NotEqual
	case *v1.Filter_ReEqual:
		re, err := regexp.New(string(e.ReEqual))
		if err != nil {
			return nil, err
		}
		o.Re = re
		o.Op = ReMatch
	case *v1.Filter_ReNotEqual:
		re, err := regexp.New(string(e.ReNotEqual))
		if err != nil {
			return nil, err
		}
		o.Re = re
		o.Op = ReNotMatch
	}
	return &o, nil
}
