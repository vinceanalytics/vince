package timeseries

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/vinceanalytics/vince/config"
	"github.com/vinceanalytics/vince/pkg/log"
)

func Open(ctx context.Context, o *config.Options) (context.Context, io.Closer, error) {
	dir := filepath.Join(o.DataPath, "ts")
	mike, err := openMike(ctx, filepath.Join(dir, "mike"))
	if err != nil {
		return nil, nil, err
	}
	unique, err := open(ctx, filepath.Join(dir, "unique"))
	if err != nil {
		mike.Close()
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, mikeKey{}, mike)
	ctx = context.WithValue(ctx, uniqueKey{}, unique)
	ctx = SetMap(ctx, NewMap())

	resource := resourceList{mike, unique}
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

func openMike(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	return badger.OpenManaged(o)
}

func open(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	return badger.Open(o)
}

type mikeKey struct{}

type uniqueKey struct{}

func GetMike(ctx context.Context) *badger.DB {
	return ctx.Value(mikeKey{}).(*badger.DB)
}

func GetUnique(ctx context.Context) *badger.DB {
	return ctx.Value(uniqueKey{}).(*badger.DB)
}
