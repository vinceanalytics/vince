package gate

import (
	"context"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/timeseries"
)

func Check(ctx context.Context, domain string) (*timeseries.Buffer, bool) {
	site, ok := caches.Site(ctx).Get(domain)
	if !ok {
		return nil, false
	}
	x := site.(*models.CachedSite)
	return timeseries.GetMap(ctx).Get(ctx,
		x.UserID, config.Get(ctx).FlushInterval.AsDuration()), true
}
