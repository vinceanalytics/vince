package timeseries

import "context"

func (t *Tables) CurrentVisitors(ctx context.Context, query Query) (*Record, error) {
	return t.events.QueryRealtime(query.Select("user_id"))
}
