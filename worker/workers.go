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
	log.Get(ctx).Debug().Str("worker", "sites_to_domain_cache").Msg("started")
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
