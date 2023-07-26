package worker

import (
	"context"
	"time"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/health"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/log"
)

// SiteCacheUpdate updates cache of active sites.
func SiteCacheUpdate(ctx context.Context, interval time.Duration) {
	setSite := caches.SetSite(ctx, interval)
	models.QuerySitesToCache(ctx, setSite)
}

func Periodic(
	ctx context.Context,
	ping *health.Ping,
	interval time.Duration,
	warm bool, // if true calls work before starting the loop
	work func(context.Context, time.Duration)) func() error {
	tick := time.NewTicker(interval)
	log.Get().Debug().Str("worker", ping.Key).
		Str("every", interval.String()).
		Msg("started")
	if warm {
		work(ctx, interval)
	}
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
	timeseries.Save(ctx)
}

func System(ctx context.Context, interval time.Duration) {
}
