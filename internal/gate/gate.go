package gate

import (
	"context"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

// Check ensure that domain is part of registered sites. Returns *timeseries.Buffer
// that collects events for domain owner.
//
// This applies rate limits configured per site/domain
func Check(ctx context.Context, domain string) (*timeseries.Buffer, bool) {
	ok := caches.AllowSite(ctx, domain)
	if !ok {
		return nil, false
	}
	return timeseries.GetMap(ctx).Get(domain), true
}
