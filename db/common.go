package db

import (
	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
)

func Timestamps(r arrow.Record) (lo, hi uint64) {
	for i := 0; i < int(r.NumCols()); i++ {
		if r.ColumnName(i) == "TimeUnixNano" {
			ts := r.Column(i).(*array.Uint64)
			lo = ts.Value(0)
			hi = ts.Value(ts.Len() - 1)
			return
		}
	}
	return
}
