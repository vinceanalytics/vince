package g

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type gKey struct{}

func Open(ctx context.Context) context.Context {
	return context.WithValue(ctx, gKey{}, &errgroup.Group{})
}

func Get(ctx context.Context) *errgroup.Group {
	return ctx.Value(gKey{}).(*errgroup.Group)
}
