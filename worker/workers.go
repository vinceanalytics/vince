package worker

import (
	"context"
	"sync"
	"time"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/system"
	"github.com/gernest/vince/timeseries"
	"github.com/gernest/vince/timex"
)

func UpdateCacheSites(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
	wg.Add(1)
	h := health.NewPing("sites_to_domain_cache")
	go updateCachedSites(ctx, wg, h.Channel, exit)
	return h
}

type cacheUpdater struct {
}

func doSiteCacheUpdate(ctx context.Context, fn func(*models.CachedSite)) {
	start := time.Now()
	defer system.SiteCacheDuration.UpdateDuration(start)
	system.SitesInCache.Set(
		models.QuerySitesToCache(ctx, fn),
	)
}

func updateCachedSites(ctx context.Context, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "sites_to_domain_cache").
		Msg("started")
	defer wg.Done()
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
			return
		case pong := <-ch:
			pong()
		case <-tick.C:
			doSiteCacheUpdate(ctx, setSite)
		}
	}
}

func LogRotate(b *log.Rotate) func(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
	return func(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
		wg.Add(1)
		h := health.NewPing("log_rotation")
		go rotateLog(ctx, b, wg, h.Channel, exit)
		return h
	}
}

func rotateLog(ctx context.Context, b *log.Rotate, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "log_rotation").Msg("started")
	defer wg.Done()
	tick := time.NewTicker(config.Get(ctx).Intervals.LogRotationCheckInterval.AsDuration())
	date := timex.Date(time.Now())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
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

func SaveTimeseries(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
	wg.Add(1)
	h := health.NewPing("timeseries_writer")
	go saveBuffer(ctx, wg, h.Channel, exit)
	return h
}

func saveBuffer(ctx context.Context, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "timeseries_writer").Msg("started")
	defer wg.Done()
	// Do 1 second  interval flushing of buffered logs
	tick := time.NewTicker(config.Get(ctx).Intervals.SaveTimeseriesBufferInterval.AsDuration())
	m := timeseries.GetMap(ctx)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case pong := <-ch:
			pong()
		case <-tick.C:
			m.Save(ctx)
		}
	}
}

func MergeTimeseries(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
	wg.Add(1)
	h := health.NewPing("timeseries_merger")
	go mergeTs(ctx, wg, h.Channel, exit)
	return h
}

func mergeTs(ctx context.Context, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "timeseries_merger").Msg("started")
	defer wg.Done()
	tick := time.NewTicker(config.Get(ctx).Intervals.MergeTimeseriesInterval.AsDuration())
	defer tick.Stop()
	var since uint64
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		case pong := <-ch:
			pong()
		case <-tick.C:
			since, err = timeseries.Merge(ctx, since, timeseries.Save)
			if err != nil {
				log.Get(ctx).Err(err).Msg("failed to merge ts")
			}
		}
	}
}

func CollectSYstemMetrics(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
	wg.Add(1)
	h := health.NewPing("system_metrics_collector")
	go collectSystemMetrics(ctx, wg, h.Channel, exit)
	return h
}

func collectSystemMetrics(ctx context.Context, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "system_metrics_collector").Msg("started")
	defer wg.Done()
	tick := time.NewTicker(config.Get(ctx).Intervals.SystemScrapeInterval.AsDuration())
	defer tick.Stop()

	// By default  we collect 24 hour windows into their own file.
	persist := time.NewTicker(24 * time.Hour)
	defer persist.Stop()

	sys := timeseries.GetSystem(ctx)
	collect := sys.Collect(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case pong := <-ch:
			pong()
		case <-persist.C:
			err := sys.Save()
			if err != nil {
				log.Get(ctx).Err(err).Msg("failed to save system metrics")
			}
		case ts := <-tick.C:
			collect(system.Collect(ts))
		}
	}
}
