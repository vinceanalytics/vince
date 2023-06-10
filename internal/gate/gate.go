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
func Check(ctx context.Context, domain string) (*timeseries.Buffer, uint64, uint64, bool) {
	uid, sid, ok := caches.AllowSite(ctx, domain)
	if !ok {
		return nil, 0, 0, false
	}
	return timeseries.GetMap(ctx).Get(uid, sid), uid, sid, true
}
