package caches

import (
	"context"

	"github.com/dgraph-io/ristretto"
)

type sessionKey struct{}

func Open(ctx context.Context) (context.Context, error) {
	session, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, sessionKey{}, session)
	return ctx, nil
}

func Close(ctx context.Context) {
	Session(ctx).Close()
}

func Session(ctx context.Context) *ristretto.Cache {
	return ctx.Value(sessionKey{}).(*ristretto.Cache)
}
