package timeseries

import "context"

func (t *Tables) CurrentVisitors(ctx context.Context, query Query) (*Record, error) {
	return t.events.Query(query)
}
