package gate

import (
	"context"

	"github.com/gernest/vince/limit"
	"github.com/gernest/vince/timeseries"
)

// Check ensure that domain is part of registered sites. Returns *timeseries.Buffer
// that collects events for domain owner.
//
// This applies rate limits configured per site/domain
func Check(ctx context.Context, domain string) (*timeseries.Buffer, bool) {
	x, ok := limit.SITES.Allow(domain)
	if !ok {
		return nil, false
	}
	return timeseries.GetMap(ctx).Get(ctx,
		x.UserID, x.ID), true
}
