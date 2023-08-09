package be

import (
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/compute/exprs"
	"github.com/substrait-io/substrait-go/proto"
	"github.com/substrait-io/substrait-go/types"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
)

var extSet = exprs.NewDefaultExtensionSet()

var all []string

var Base = func() (r types.NamedStruct) {
	f := entry.Fields()
	r.Struct.Nullability = proto.Type_NULLABILITY_REQUIRED
	for i := range f {
		x := &f[i]
		r.Names = append(r.Names, x.Name)
		all = append(all, x.Name)
		switch x.Type.ID() {
		case arrow.STRING:
			r.Struct.Types = append(r.Struct.Types, &types.StringType{
				Nullability: proto.Type_NULLABILITY_REQUIRED,
			})
		case arrow.INT64:
			r.Struct.Types = append(r.Struct.Types, &types.Int64Type{
				Nullability: proto.Type_NULLABILITY_REQUIRED,
			})
		default:
			must.AssertFMT(false)("unsupported field type %v", x.Type.ID())
		}
	}
	return
}()

func Schema(fields ...string) (r types.NamedStruct) {
	m := make(map[string]struct{})
	for i := range fields {
		m[fields[i]] = struct{}{}
	}
	r.Struct.Nullability = proto.Type_NULLABILITY_REQUIRED
	for i := range Base.Names {
		_, ok := m[Base.Names[i]]
		if !ok {
			continue
		}
		r.Struct.Types = append(r.Struct.Types,
			Base.Struct.Types[i].WithNullability(proto.Type_NULLABILITY_REQUIRED))
	}
	return
}
