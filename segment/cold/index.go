package cold

import (
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/vinceanalytics/vince/camel"
	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/segment"
	"github.com/vinceanalytics/vince/segment/ice"
)

var skipFields = map[string]struct{}{
	camel.Case(v1.Filters_ID.String()):       {},
	camel.Case(v1.Filters_Session.String()):  {},
	camel.Case(v1.Filters_Bounce.String()):   {},
	camel.Case(v1.Filters_Duration.String()): {},
}

func New(r arrow.Record) (segment.Segment, error) {
	document := NewRecord(r, func(r arrow.Record, i int) bool {
		_, ok := skipFields[r.ColumnName(i)]
		return !ok
	})
	defer document.Release()
	seg, _, err := ice.New([]segment.Document{document}, func(s string, i int) float32 { return 0 })
	return seg, err
}
