package limit

import (
	"context"
	"time"

	"github.com/dgraph-io/ristretto"
	"golang.org/x/time/rate"
)

type Rate struct {
	cache *ristretto.Cache
}

func New(cache *ristretto.Cache) *Rate {
	return &Rate{cache: cache}
}

func (r *Rate) Allow(id uint64, events uint, duration time.Duration) bool {
	l, ok := r.cache.Get(id)
	if !ok {
		x := rate.NewLimiter(rate.Limit(float64(events)/duration.Seconds()), 10)
		r.cache.Set(id, x, 1)
		return x.Allow()
	}
	return l.(*rate.Limiter).Allow()
}

func (r *Rate) Close() {
	r.cache.Close()
}

type rateLimitKey struct{}

func Set(ctx context.Context, r *Rate) context.Context {
	return context.WithValue(ctx, rateLimitKey{}, r)
}

func Get(ctx context.Context) *Rate {
	return ctx.Value(rateLimitKey{}).(*Rate)
}
