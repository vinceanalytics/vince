package timeseries

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Open(ctx context.Context, o *config.Options) (context.Context, io.Closer, error) {
	dir := filepath.Join(o.DataPath, "ts")
	temporary, err := managed(ctx, filepath.Join(dir, "temp"))
	if err != nil {
		return nil, nil, err
	}
	unique, err := open(ctx, filepath.Join(dir, "unique"))
	if err != nil {
		temporary.Close()
		return nil, nil, err
	}
	permanent, err := managed(ctx, filepath.Join(dir, "permanent"))
	if err != nil {
		temporary.Close()
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, temporaryKey{}, temporary)
	ctx = context.WithValue(ctx, uniqueKey{}, unique)
	ctx = context.WithValue(ctx, permanentKey{}, permanent)
	m := NewMap(o.Intervals.TSSync)
	ctx = SetMap(ctx, m)

	resource := resourceList{temporary, unique, permanent, m}
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

func managed(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithCompression(options.ZSTD).
		WithLogger(log.Badger(ctx))
	return badger.OpenManaged(o)
}

func open(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	return badger.Open(o)
}

type temporaryKey struct{}
type uniqueKey struct{}
type permanentKey struct{}

func Temporary(ctx context.Context) *badger.DB {
	return ctx.Value(temporaryKey{}).(*badger.DB)
}

func GetUnique(ctx context.Context) *badger.DB {
	return ctx.Value(uniqueKey{}).(*badger.DB)
}

func Permanent(ctx context.Context) *badger.DB {
	return ctx.Value(permanentKey{}).(*badger.DB)
}
