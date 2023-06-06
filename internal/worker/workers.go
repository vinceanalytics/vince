package worker

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/log"
)

// SiteCacheUpdate updates cache of active sites.
func SiteCacheUpdate(ctx context.Context, interval time.Duration) {
	start := core.Now(ctx)
	defer system.SiteCacheDuration.Observe(time.Since(start).Seconds())
	setSite := caches.SetSite(ctx, interval)
	system.SitesInCache.Set(
		models.QuerySitesToCache(ctx, setSite),
	)
}

func Periodic(
	ctx context.Context,
	ping *health.Ping,
	interval time.Duration,
	work func(context.Context, time.Duration)) func() error {
	tick := time.NewTicker(interval)
	log.Get().Debug().Str("worker", ping.Key).
		Str("every", interval.String()).
		Msg("started")
	return func() error {
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return nil
			case pong := <-ping.Channel:
				pong()
			case <-tick.C:
				work(ctx, interval)
			}
		}
	}
}

// SaveBuffers persists collected Entry Buffers to the timeseries storage.
func SaveBuffers(ctx context.Context, interval time.Duration) {
	timeseries.GetMap(ctx).Save(ctx)
}
