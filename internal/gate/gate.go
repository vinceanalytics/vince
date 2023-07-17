package gate

import (
	"context"

	"github.com/vinceanalytics/vince/internal/caches"
)

// Check ensure that domain is part of registered sites. Returns *timeseries.Buffer
// that collects events for domain owner.
//
// This applies rate limits configured per site/domain
func Check(ctx context.Context, domain string) bool {
	return caches.AllowSite(ctx, domain)
}
