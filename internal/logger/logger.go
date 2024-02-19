package logger

import (
	"context"
	"log/slog"
	"os"
)

func Fail(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

type key struct{}

func With(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, log)
}

func Get(ctx context.Context) *slog.Logger {
	return ctx.Value(key{}).(*slog.Logger)
}
