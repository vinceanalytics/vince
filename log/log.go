package log

import (
	"context"

	"github.com/rs/zerolog"
)

type loggerKey struct{}

func Set(ctx context.Context, lg *zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, lg)
}

func Get(ctx context.Context) *zerolog.Logger {
	return ctx.Value(loggerKey{}).(*zerolog.Logger)
}
