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
	if p := ctx.Value(paramKey{}); p != nil {
		return p.(Params)
	}
	return nil
}

func (p Params) Get(k string) string {
	if p == nil {
		return ""
	}
	return p[k]
}

func Re(re *regexp2.Match) Params {
	p := make(Params)
	for _, g := range re.Groups()[1:] {
		p[g.Name] = g.Captures[0].String()
	}
	return p
}
