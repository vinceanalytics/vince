package timeseries

import (
	"context"

	"github.com/vinceanalytics/vince/internal/api/aggregates"
	"github.com/vinceanalytics/vince/internal/fieldset"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/util/xtime"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func Timeseries(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query, metrics []string) (map[string]*aggregates.Stats, error) {
	values := make(map[string]*aggregates.Stats)
	fields := fieldset.From(metrics...)

	format := params.Interval().Format()
	err := ts.Select(ctx, domain, params.Start(), params.End(), params.Interval(), params.Filter(), func(shard, view uint64, columns *roaring.Bitmap) error {
		timestamp := xtime.UnixMilli(int64(view)).Format(format)
		m, ok := values[timestamp]
		if !ok {
			m = aggregates.NewStats(fields)
			values[timestamp] = m
		}
		return m.Read(ctx, ts, shard, view, columns, fields)
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}
