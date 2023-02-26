package caches

import (
	"context"

	"github.com/dgraph-io/ristretto"
)

type rateKey struct{}
type sessionKey struct{}

type Hooks struct {
	Rate    Hook
	Session Hook
}

type Hook struct {
	OnEvict, OnReject func(item *ristretto.Item)
	OnExit            func(any)
}

func Open(ctx context.Context, hooks Hooks) (context.Context, error) {
	session, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
		OnEvict:     hooks.Session.OnEvict,
		OnReject:    hooks.Session.OnReject,
		OnExit:      hooks.Session.OnExit,
	})
	if err != nil {
		return nil, err
	}

	// we use item based cost for rate limiting. We don't want to have too many rate limiters
	// in memory for less active sites. A single RateLimiter object is of size 80 bytes
	// Setting a maximum cost of  10 MB gives us around 130k active rate limit session
	// before eviction.
	rate, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     10 << 20,
		BufferItems: 64,
		OnEvict:     hooks.Rate.OnEvict,
		OnReject:    hooks.Rate.OnReject,
		OnExit:      hooks.Rate.OnExit,
	})
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, sessionKey{}, session)
	ctx = context.WithValue(ctx, rateKey{}, rate)
	return ctx, nil
}

func Close(ctx context.Context) {
	Session(ctx).Close()
	Rate(ctx).Close()
}

func Session(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sessionKey{}).(*ristretto.Cache)
}

func Rate(ctx context.Context) *ristretto.Cache {
	return ctx.Value(rateKey{}).(*ristretto.Cache)
}
