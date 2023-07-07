package timeseries

import (
	"context"

	"github.com/vinceanalytics/vince/pkg/spec"
)

func QuerySeries(ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[[]uint64]) {
	return
}

func QueryAggregate(ctx context.Context, uid, sid uint64, o spec.QueryPropertyOptions) (result spec.PropertyResult[uint64]) {
	return
}

func Stat(ctx context.Context, uid, sid uint64, metric spec.Metric) (o spec.Global[uint64]) {
	return
}

func Stats(ctx context.Context, uid, sid uint64) (o spec.Global[spec.Metrics]) {
	return
}

func GlobalAggregate(ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.Series[uint64]) {
	return
}

func GlobalSeries(ctx context.Context, uid, sid uint64, o spec.QueryOptions) (r spec.Series[[]uint64]) {
	return
}
