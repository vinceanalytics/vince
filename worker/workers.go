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
)

func UpdateCacheSites(ctx context.Context, wg *sync.WaitGroup, exit func()) *health.Ping {
	wg.Add(1)
	h := health.NewPing("sites_to_domain_cache")
	go updateCachedSites(ctx, wg, h.Channel, exit)
	return h
}

func updateCachedSites(ctx context.Context, wg *sync.WaitGroup, ch health.PingChannel, exit func()) {
	log.Get(ctx).Debug().Str("worker", "sites_to_domain_cache").Msg("started")
	defer wg.Done()
	sites := make([]*models.CachedSite, 0, 4098)
	err := models.QuerySitesToCache(models.Get(ctx), &sites)
	if err != nil {
		log.Get(ctx).Err(err).Str("worker", "sites_to_domain_cache").Msg("failed querying sites to cache")
		exit()
		return
	}
	cache := caches.Site(ctx)
	cache.Clear()
	for _, s := range sites {
		cache.Set(s.Domain, s, 1)
	}
	interval := config.Get(ctx).SitesByDomainCacheRefreshInterval
	tick := time.NewTicker(interval.AsDuration())
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case pong := <-ch:
			pong()
		case <-tick.C:
			sites = sites[:0]
			err := models.QuerySitesToCache(models.Get(ctx), &sites)
			if err != nil {
				log.Get(ctx).Err(err).Str("worker", "sites_to_domain_cache").Msg("failed querying sites to cache")
				exit()
				return
			}
			models.SitesMu.Lock()
			cache.Clear()
			for _, s := range sites {
				cache.Set(s.Domain, s, 1)
			}
			models.SitesMu.Unlock()
		}
	}

}
