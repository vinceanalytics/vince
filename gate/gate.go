package gate

import (
	"context"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/limit"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/timeseries"
)

// Check ensure that domain is part of registered sites. Returns *timeseries.Buffer
// that collects events for domain owner.
//
// This applies rate limits configured per site/domain
func Check(ctx context.Context, domain string) (*timeseries.Buffer, bool) {
	conf := config.Get(ctx)
	if conf.Env == config.Config_load {
		// special case
		return timeseries.GetMap(ctx).Get(ctx,
			0, 0), true
	}
	site, ok := caches.Site(ctx).Get(domain)
	if !ok {
		return nil, false
	}
	x := site.(*models.CachedSite)
	ok = limit.SITES.Allow(x.RateLimit())
	if !ok {
		return nil, false
	}
	return timeseries.GetMap(ctx).Get(ctx,
		x.UserID, x.ID), true
}
