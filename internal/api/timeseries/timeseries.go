package timeseries

import (
	"context"

	"github.com/vinceanalytics/vince/internal/api/aggregates"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/ro2"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/internal/util/xtime"
	"github.com/vinceanalytics/vince/internal/web/query"
)

func Timeseries(ctx context.Context, ts *timeseries.Timeseries, domain string, params *query.Query, metrics []string) (map[string]*aggregates.Stats, error) {
	values := make(map[string]*aggregates.Stats)
	fields := models.DataForMetrics(metrics...)

	format := params.Interval().Format()
	err := ts.Select(ctx, fields, domain, params.Start(), params.End(), params.Interval(), params.Filter(), func(ra *timeseries.Cursor, dataField models.Field, view, shard uint64, columns *ro2.Bitmap) error {
		timestamp := xtime.UnixMilli(int64(view)).Format(format)
		m, ok := values[timestamp]
		if !ok {
			m = new(aggregates.Stats)
			values[timestamp] = m
		}
		return m.Read(ra, dataField, view, shard, columns)
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}
