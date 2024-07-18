package guard

import (
	"context"
	"sync"

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

func New(limit float64, domains []string) *BasicGuard {
	b := &BasicGuard{
		domains: make(map[string]struct{}),
		rate:    rate.NewLimiter(rate.Limit(limit), 0),
	}
	for _, d := range domains {
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
