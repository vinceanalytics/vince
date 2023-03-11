package worker

import (
	"context"
	"runtime/trace"
	"sync"
	"time"

	"github.com/gernest/vince/caches"
	"github.com/gernest/vince/config"
	"github.com/gernest/vince/health"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
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
	sites []*models.CachedSite
	ttl   time.Duration
}

// Do updates the cache with new *models.CachedSite entries
func (c *cacheUpdater) Do(ctx context.Context) {
	ctx, task := trace.NewTask(ctx, "sites_to_domain_cache_update")
	defer task.End()
	c.sites = c.sites[:0]
	err := models.QuerySitesToCache(models.Get(ctx), &c.sites)
	if err != nil {
		log.Get(ctx).Err(err).Str("worker", "sites_to_domain_cache").Msg("failed querying sites to cache")
		return
	}
	cache := caches.Site(ctx)
	for _, s := range c.sites {
		cache.SetWithTTL(s.Domain, s, 1, c.ttl)
	}
}

func updateCachedSites(ctx context.Context, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "sites_to_domain_cache").
		Int("some", 1).
		Int("some2", 2).
		Msg("started")
	defer wg.Done()
	interval := config.Get(ctx).SitesByDomainCacheRefreshInterval
	work := &cacheUpdater{
		sites: make([]*models.CachedSite, 0, 4098),
		ttl:   interval.AsDuration(),
	}
	// On startup , fill the cache first before the next interval. Ensures we are
	// operational  on the get go.
	work.Do(ctx)
	tick := time.NewTicker(interval.AsDuration())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case pong := <-ch:
			pong()
		case <-tick.C:
			work.Do(ctx)
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
	// Do 1 second  interval flushing of buffered logs
	tick := time.NewTicker(time.Second)
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
	tick := time.NewTicker(config.Get(ctx).FlushInterval.AsDuration())
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
