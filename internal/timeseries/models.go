package timeseries

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Open(ctx context.Context, o *config.Options) (context.Context, io.Closer, error) {
	dir := filepath.Join(o.DataPath, "ts")
	neo, err := badger.Open(badger.DefaultOptions(dir).
		WithLogger(badgerLogger{}).
		WithCompression(options.ZSTD))
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, storeKey{}, neo)
	m := NewMap(o.Intervals.TSSync)
	ctx = SetMap(ctx, m)
	resource := resourceList{neo, m}
	return ctx, resource, nil
}

type resourceList []io.Closer

func (r resourceList) Close() error {
	err := make([]error, len(r))
	for i := range err {
		err[i] = r[i].Close()
	}
	return errors.Join(err...)
}

type storeKey struct{}

func Store(ctx context.Context) *badger.DB {
	return ctx.Value(storeKey{}).(*badger.DB)
}

func Save(ctx context.Context, b *Buffer) {
	err := b.neo.Save(Store(ctx))
	if err != nil {
		log.Get().Err(err).
			Str("domain", b.domain).
			Msg("failed saving buffer")
	}
	b.Release()
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
