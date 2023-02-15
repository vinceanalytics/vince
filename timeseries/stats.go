package timeseries

import "context"

func (t *Tables) CurrentVisitors(ctx context.Context, query Query) (*Record, error) {
	return t.Events.Query(ctx, query.Select("user_id", "referrer").
		Filter("domain", Eq("vince.test")).Filter("referrer", Eq("vince.test")))
}
