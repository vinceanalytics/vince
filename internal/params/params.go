package params

import (
	"context"

	"github.com/dlclark/regexp2"
)

type Params map[string]string

type paramKey struct{}

func Set(ctx context.Context, p Params) context.Context {
	return context.WithValue(ctx, paramKey{}, p)
}

func Get(ctx context.Context) Params {
	return ctx.Value(paramKey{}).(Params)
}

func Re(re *regexp2.Match) Params {
	p := make(Params)
	for _, g := range re.Groups()[1:] {
		p[g.Name] = g.Captures[0].String()
	}
	return p
}
