package be

import (
	"time"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/apache/arrow/go/v13/arrow/array"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/entry"
)

var base = must.Must(time.Parse(time.RFC822Z, time.RFC822Z))().UTC()

func TimeBucket(a *array.Int64, interval time.Duration) arrow.Array {
	b := array.NewInt64Builder(entry.Pool)
	defer b.Release()
	values := a.Int64Values()
	at := startingPoint(
		time.UnixMilli(values[0]),
		interval,
	).UnixMilli()
	for i := range values {
		if values[i] >= at {
			at = time.UnixMilli(at).Add(interval).UnixMilli()
		}
		b.UnsafeAppend(at)
	}
	return b.NewArray()
}

// Calculates the first bucket that value a will fall into
func startingPoint(a time.Time, i time.Duration) time.Time {
	n := a.Sub(base)
	x := n / i
	return base.Add(i * x)
}
