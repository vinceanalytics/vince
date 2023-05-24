package config

import (
	"context"
	"fmt"
)

type configKey struct{}

func Get(ctx context.Context) *Options {
	return ctx.Value(configKey{}).(*Options)
}

func Load(base *Options) (context.Context, error) {
	sec, a, err := setupSecrets(base)
	if err != nil {
		return nil, fmt.Errorf("failed to setup secrets %v", err)
	}
	baseCtx := context.WithValue(context.Background(), configKey{}, base)
	baseCtx = context.WithValue(baseCtx, securityKey{}, sec)
	baseCtx = context.WithValue(baseCtx, ageKey{}, a)
	return baseCtx, nil
}
