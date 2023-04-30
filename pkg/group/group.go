package group

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type groupKey struct{}

func Set(ctx context.Context, g *errgroup.Group) context.Context {
	return context.WithValue(ctx, groupKey{}, g)
}

func GO(ctx context.Context, f func(context.Context) error) {
	ctx.Value(groupKey{}).(*errgroup.Group).Go(func() error {
		return f(ctx)
	})
}
