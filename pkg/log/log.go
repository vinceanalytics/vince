package log

import (
	"context"

	"github.com/rs/zerolog"
)

func Get(ctx context.Context) *zerolog.Logger {
	return zerolog.Ctx(ctx)
}
