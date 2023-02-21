package params

import (
	"context"
	"regexp"
)

type Params map[string]string

type paramKey struct{}

func Set(ctx context.Context, p Params) context.Context {
	return context.WithValue(ctx, paramKey{}, p)
}

func Get(ctx context.Context) Params {
	return ctx.Value(paramKey{}).(Params)
}

func Re(re *regexp.Regexp, path string) Params {
	m := re.FindStringSubmatch(path)
	p := make(Params)
	for k, v := range re.SubexpNames() {
		p[v] = m[k]
	}
	return p
}
