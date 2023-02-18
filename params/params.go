package params

import "context"

type Params map[string]string

type paramKey struct{}

func Set(ctx context.Context, p Params) context.Context {
	return context.WithValue(ctx, paramKey{}, p)
}

func Get(ctx context.Context) Params {
	return ctx.Value(paramKey{}).(Params)
}
