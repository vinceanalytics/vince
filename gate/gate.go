package gate

import (
	"context"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/timeseries"
)

// Check ensure that domain is part of registered sites. Returns *timeseries.Buffer
// that collects events for domain owner.
//
// This applies rate limits configured per site/domain
func Check(ctx context.Context, domain string) (*timeseries.Buffer, bool) {
	uid, sid, ok := caches.AllowSite(ctx, domain)
	if !ok {
		return nil, false
	}
	return timeseries.GetMap(ctx).Get(ctx, uid, sid), true
}
