package timeseries

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/neo"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Open(ctx context.Context, o *config.Options) (context.Context, io.Closer, error) {
	dir := filepath.Join(o.DataPath, "ts")
	db, err := badger.Open(badger.DefaultOptions(filepath.Join(dir, "series")).
		WithLogger(badgerLogger{}).
		WithCompression(options.ZSTD))
	if err != nil {
		return nil, nil, err
	}
	a := neo.NewBlock(dir, db)
	ctx = context.WithValue(ctx, storeKey{}, db)
	ctx = context.WithValue(ctx, blockKey{}, a)
	resource := resourceList{a, db}
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
type blockKey struct{}

func Store(ctx context.Context) *badger.DB {
	return ctx.Value(storeKey{}).(*badger.DB)
}

func Block(ctx context.Context) *neo.ActiveBlock {
	return ctx.Value(blockKey{}).(*neo.ActiveBlock)
}

func Save(ctx context.Context) {
	err := Block(ctx).Save(ctx)
	if err != nil {
		log.Get().Fatal().Err(err).Msg("failed to save block")
	}
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
