package db

import (
	"github.com/gernest/rbf/dsl/mutex"
	"github.com/gernest/rbf/dsl/query"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func filterProperties(fs ...*v1.Filter) query.Filter {
	if len(fs) == 0 {
		return query.Noop{}
	}
	ls := make(query.And, len(fs))
	for i := range fs {
		ls[i] = filterProperty(fs[i])
	}
	return ls
}

func filterProperty(f *v1.Filter) query.Filter {
	var o mutex.OP
	switch f.Op {
	case v1.Filter_equal:
		o = mutex.EQ
	case v1.Filter_not_equal:
		o = mutex.NEQ
	case v1.Filter_re_equal:
		o = mutex.RE
	case v1.Filter_re_not_equal:
		o = mutex.NRE
	default:
		return query.Noop{}
	}
	return &mutex.MatchString{
		Field: f.Property.String(),
		Op:    o,
		Value: f.Value,
	}
}
