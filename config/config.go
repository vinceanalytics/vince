package config

import (
	"context"
	"net/mail"
)

type Config struct {
	Mailer Mailer
}

type Mailer struct {
	Address mail.Address
}

type configKey struct{}

func Set(ctx context.Context, conf *Config) context.Context {
	return context.WithValue(ctx, configKey{}, conf)
}

func Get(ctx context.Context) *Config {
	return ctx.Value(configKey{}).(*Config)
}
