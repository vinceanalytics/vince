package guard

import (
	"context"
	"sync"

	v1 "github.com/vinceanalytics/staples/staples/gen/go/staples/v1"
	"golang.org/x/time/rate"
)

type Guard interface {
	Allow() bool
	Accept(domain string) bool
}

type guardKey struct{}

func With(ctx context.Context, g Guard) context.Context {
	return context.WithValue(ctx, guardKey{}, g)
}

func Get(ctx context.Context) Guard {
	return ctx.Value(guardKey{}).(Guard)
}

type BasicGuard struct {
	domains map[string]struct{}
	rate    *rate.Limiter
	mu      sync.Mutex
}

func New(c *v1.Config) *BasicGuard {
	b := &BasicGuard{
		domains: make(map[string]struct{}),
		rate:    rate.NewLimiter(rate.Limit(c.RateLimit), 0),
	}
	for _, d := range c.Domains {
		b.domains[d] = struct{}{}
	}
	return b
}

func (b *BasicGuard) Allow() (ok bool) {
	b.mu.Lock()
	ok = b.rate.Allow()
	b.mu.Unlock()
	return
}

func (b *BasicGuard) Accept(domain string) bool {
	_, ok := b.domains[domain]
	return ok
}
