package log

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog"
)

var _ badger.Logger = (*badgerLogger)(nil)

type badgerLogger struct {
	lg zerolog.Logger
}

func Badger(ctx context.Context) badger.Logger {
	return &badgerLogger{lg: Get().Level(zerolog.WarnLevel)}
}

func (b *badgerLogger) Errorf(format string, args ...interface{}) {
	b.lg.Error().Msgf(format, args...)
}
func (b *badgerLogger) Warningf(format string, args ...interface{}) {
	b.lg.Warn().Msgf(format, args...)
}

func (b *badgerLogger) Infof(format string, args ...interface{}) {
	b.lg.Info().Msgf(format, args...)
}

func (b *badgerLogger) Debugf(format string, args ...interface{}) {
	b.lg.Debug().Msgf(format, args...)
}
