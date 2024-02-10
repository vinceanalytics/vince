package cold

import (
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/vinceanalytics/vince/segment"
)

type record struct {
	r arrow.Record
}

var _ segment.Document = (*record)(nil)

func (r *record) Analyze() {}

func (r *record) EachField(vf segment.VisitField) {
	for i := 0; i < int(r.r.NumCols()); i++ {
		_, ok := skipFields[r.r.ColumnName(i)]
		if ok {
			continue
		}
		visitArray(r.r.Column(i).(*array.Dictionary), r.r.ColumnName(i), vf)
	}
}

func visitArray(a *array.Dictionary, name string, cb segment.VisitField) {
	cb(&DictionaryField{
		base: base{
			name: name,
			len:  a.Len(),
		},
		array: a,
		data:  a.Dictionary().(*array.Binary),
	})
}
