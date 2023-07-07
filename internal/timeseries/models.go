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

	unique, err := open(ctx, filepath.Join(dir, "unique"))
	if err != nil {
		return nil, nil, err
	}

	neo, err := OpenStore(ctx, dir)
	if err != nil {
		unique.Close()
		return nil, nil, err
	}

	ctx = context.WithValue(ctx, uniqueKey{}, unique)
	ctx = context.WithValue(ctx, storeKey{}, neo)
	m := NewMap(o.Intervals.TSSync)
	ctx = SetMap(ctx, m)

	resource := resourceList{unique, neo, m}
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

func open(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	return badger.Open(o)
}

type uniqueKey struct{}

type storeKey struct{}

func Unique(ctx context.Context) *badger.DB {
	return ctx.Value(uniqueKey{}).(*badger.DB)
}

func Store(ctx context.Context) *V9 {
	return ctx.Value(storeKey{}).(*V9)
}

func GC(ctx context.Context) {
	Unique(ctx).RunValueLogGC(0.5)
}
