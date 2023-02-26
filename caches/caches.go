package caches

import (
	"context"

	"github.com/dgraph-io/ristretto"
)

type sessionKey struct{}
type sitesKey struct{}

func Open(ctx context.Context) (context.Context, error) {
	session, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	sites, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     10 << 20,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, sessionKey{}, session)
	ctx = context.WithValue(ctx, sitesKey{}, sites)
	return ctx, nil
}

func Close(ctx context.Context) {
	Session(ctx).Close()
	Site(ctx).Close()
}

func Session(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sessionKey{}).(*ristretto.Cache)
}

func Site(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sitesKey{}).(*ristretto.Cache)
}
