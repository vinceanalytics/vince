package timeseries

import (
	"context"

	"github.com/vinceanalytics/vince/internal/api/aggregates"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/roaring"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/util/xtime"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func Timeseries(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query, metrics []string) (map[string]*aggregates.Stats, error) {
	values := make(map[string]*aggregates.Stats)
	fields := models.DataForMetrics(metrics...)

	format := params.Interval().Format()
	ts.Select(ctx, fields, domain, params.Start(), params.End(), params.Interval(), params.Filter(), func(dataField models.Field, view, shard uint64, columns, ra *roaring.Bitmap) bool {
		timestamp := xtime.UnixMilli(int64(view)).Format(format)
		m, ok := values[timestamp]
		if !ok {
			m = new(aggregates.Stats)
			values[timestamp] = m
		}
		m.Read(dataField, view, shard, columns, ra)
		return true
	})
	return values, nil
}
