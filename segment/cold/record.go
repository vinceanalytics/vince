package cold

import (
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/vinceanalytics/vince/segment"
)

type Record struct {
	record arrow.Record
	accept func(r arrow.Record, idx int) bool
}

func NewRecord(r arrow.Record, accept func(arrow.Record, int) bool) *Record {
	r.Retain()
	if accept == nil {
		accept = func(r arrow.Record, i int) bool { return true }
	}
	return &Record{
		record: r,
		accept: accept,
	}
}

var _ segment.Document = (*Record)(nil)

func (r *Record) Analyze() {}

func (r *Record) Release() {
	r.record.Release()
}

func (r *Record) EachField(vf segment.VisitField) {
	for i := 0; i < int(r.record.NumCols()); i++ {
		visitArray(r.record.Column(i), r.record.ColumnName(i), vf)
	}
}

func visitArray(a arrow.Array, name string, cb segment.VisitField) {
	switch e := a.(type) {
	case *array.Int64:
		x := NewInt64(name, e)
		cb(x)
		x.Release()
	case *array.Dictionary:
		if v, ok := e.Dictionary().(*array.Binary); ok {
			cb(NewDict(name, e, v))
		}
	}
}
