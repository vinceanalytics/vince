package guard

import "context"

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
