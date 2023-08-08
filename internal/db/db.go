package db

import (
	"context"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/pkg/log"
)

type key struct{}

func Open(ctx context.Context, path string) (context.Context, *badger.DB) {
	dir := filepath.Join(path, "db")
	db := must.Must(badger.Open(badger.DefaultOptions(filepath.Join(dir, "series")).
		WithLogger(badgerLogger{}).
		WithCompression(options.ZSTD)))(
		"failed to open badger db for  timeseries storage dir:", dir,
	)
	return context.WithValue(ctx, key{}, db), db
}

func Get(ctx context.Context) *badger.DB {
	return ctx.Value(key{}).(*badger.DB)
}

var _ badger.Logger = (*badgerLogger)(nil)

type badgerLogger struct {
}

func (badgerLogger) Errorf(format string, args ...interface{}) {
	log.Get().Error().Msgf(format, args...)
}
func (badgerLogger) Warningf(format string, args ...interface{}) {
	log.Get().Warn().Msgf(format, args...)
}

func (badgerLogger) Infof(format string, args ...interface{}) {
	log.Get().Info().Msgf(format, args...)
}

func (b badgerLogger) Debugf(format string, args ...interface{}) {
	log.Get().Debug().Msgf(format, args...)
}
