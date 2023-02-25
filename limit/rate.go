package limit

import (
	"context"
	"time"

	"github.com/gernest/vince/caches"
	"golang.org/x/time/rate"
)

func Allow(ctx context.Context, id uint64, events uint, duration time.Duration) bool {
	cache := caches.Rate(ctx)
	l, ok := cache.Get(id)
	if !ok {
		x := rate.NewLimiter(rate.Limit(float64(events)/duration.Seconds()), 10)
		cache.Set(id, x, 1)
		return x.Allow()
	}
	return l.(*rate.Limiter).Allow()
}
