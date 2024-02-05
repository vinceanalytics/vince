package api

import (
	"context"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"github.com/vinceanalytics/staples/staples/stats"
)

var _ v1.StatsServer = (*API)(nil)

func (a *API) GetRealtimeVisitors(ctx context.Context, req *v1.Realtime_GetVisitorsRequest) (*v1.Realtime_GetVisitorsResponse, error) {
	return stats.Realtime(ctx, req)
}
func (a *API) GetAggregate(ctx context.Context, req *v1.Aggregate_GetAggegateRequest) (*v1.Aggregate_GetAggregateResponse, error) {
	return stats.Aggregate(ctx, req)
}
func (a *API) GetTimeseries(ctx context.Context, req *v1.Timeseries_GetTimeseriesRequest) (*v1.Timeseries_GetTimeseriesResponse, error) {
	return stats.TimeSeries(ctx, req)
}
func (a *API) GetBreakDown(ctx context.Context, req *v1.BreakDown_GetBreakDownRequest) (*v1.BreakDown_GetBreakDownResponse, error) {
	return stats.BreakDown(ctx, req)
}
