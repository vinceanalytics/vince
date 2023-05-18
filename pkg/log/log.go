package log

import (
	"context"

	"github.com/rs/zerolog"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func Get(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
