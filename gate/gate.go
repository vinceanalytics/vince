package gate

import (
	"context"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/limit"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/timeseries"
)

func Check(ctx context.Context, domain string) (*timeseries.Buffer, bool) {
	models.SitesMu.Lock()
	site, ok := caches.Site(ctx).Get(domain)
	models.SitesMu.Unlock()
	if !ok {
		return nil, false
	}
	x := site.(*models.CachedSite)
	ok = limit.SITES.Allow(x.RateLimit())
	if !ok {
		return nil, false
	}
	return timeseries.GetMap(ctx).Get(ctx,
		x.UserID, config.Get(ctx).FlushInterval.AsDuration()), true
}
