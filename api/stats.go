package api

import (
	"context"

	v1 "github.com/vinceanalytics/vince/gen/go/staples/v1"
	"github.com/vinceanalytics/vince/stats"
)

var _ v1.StatsServer = (*API)(nil)

func (a *API) RealtimeVisitors(ctx context.Context, req *v1.Realtime_Request) (*v1.Realtime_Response, error) {
	return stats.Realtime(ctx, req)
}
func (a *API) Aggregate(ctx context.Context, req *v1.Aggregate_Request) (*v1.Aggregate_Response, error) {
	return stats.Aggregate(ctx, req)
}
func (a *API) Timeseries(ctx context.Context, req *v1.Timeseries_Request) (*v1.Timeseries_Response, error) {
	return stats.TimeSeries(ctx, req)
}
func (a *API) BreakDown(ctx context.Context, req *v1.BreakDown_Request) (*v1.BreakDown_Response, error) {
	return stats.BreakDown(ctx, req)
}
