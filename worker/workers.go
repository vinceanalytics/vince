package worker

import (
	"context"
	"time"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/system"
	"github.com/gernest/vince/timeseries"
	"github.com/gernest/vince/timex"
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
		log.Get(ctx).Debug().Str("worker", "sites_to_domain_cache").
			Msg("started")
		interval := config.Get(ctx).Intervals.SitesByDomainCacheRefreshInterval.AsDuration()
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

func LogRotate(ctx context.Context, b *log.Rotate, f func(*health.Ping)) func() error {
	h := health.NewPing("log_rotation")
	f(h)
	return rotateLog(ctx, b, h.Channel)
}

func rotateLog(ctx context.Context, b *log.Rotate, ch health.PingChannel) func() error {
	return func() error {
		log.Get(ctx).Debug().Str("worker", "log_rotation").Msg("started")
		tick := time.NewTicker(config.Get(ctx).Intervals.LogRotationCheckInterval.AsDuration())
		date := timex.Date(time.Now())
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case pong := <-ch:
				pong()
			case x := <-tick.C:
				x = timex.Date(x)
				if !x.Equal(date) {
					// Any change on date warrants log rotation
					err := b.Rotate()
					if err != nil {
						log.Get(ctx).Err(err).Msg("failed log rotation")
					}
					date = x
				}
				b.Flush()
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
		log.Get(ctx).Debug().Str("worker", "timeseries_writer").Msg("started")
		tick := time.NewTicker(config.Get(ctx).Intervals.SaveTimeseriesBufferInterval.AsDuration())
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
