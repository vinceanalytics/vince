package worker

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/log"
)

func UpdateCacheSites(ctx context.Context, f func(*health.Ping)) func() error {
	h := health.NewPing("sites_to_domain_cache")
	f(h)
	return updateCachedSites(ctx, h.Channel)
}

func doSiteCacheUpdate(ctx context.Context, fn func(*models.CachedSite)) {
	start := time.Now()
	defer system.SiteCacheDuration.Observe(time.Since(start).Seconds())
	system.SitesInCache.Set(
		models.QuerySitesToCache(ctx, fn),
	)
}

func updateCachedSites(ctx context.Context, ch health.PingChannel) func() error {
	return func() error {
		log.Get().Debug().Str("worker", "sites_to_domain_cache").
			Msg("started")
		interval := config.Get(ctx).Intervals.SiteCache
		// On startup , fill the cache first before the next interval. Ensures we are
		// operational  on the get go.
		setSite := caches.SetSite(ctx, interval)
		doSiteCacheUpdate(ctx, setSite)
		tick := time.NewTicker(interval)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case pong := <-ch:
				pong()
			case <-tick.C:
				doSiteCacheUpdate(ctx, setSite)
			}
		}
	}
}

func SaveTimeseries(ctx context.Context, f func(*health.Ping)) func() error {
	h := health.NewPing("timeseries_writer")
	f(h)
	return saveBuffer(ctx, h.Channel)
}

func saveBuffer(ctx context.Context, ch health.PingChannel) func() error {
	return func() error {
		log.Get().Debug().Str("worker", "timeseries_writer").Msg("started")
		tick := time.NewTicker(config.Get(ctx).Intervals.TSSync)
		m := timeseries.GetMap(ctx)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case pong := <-ch:
				pong()
			case <-tick.C:
				m.Save(ctx)
			}
		}
	}
}
