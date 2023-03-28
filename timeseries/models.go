package timeseries

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/gernest/vince/log"
)

// Creates two badger.DB instances one for temporary aggregate and another for
// permanent storage/query.
func Open(ctx context.Context, dir string) (context.Context, io.Closer, error) {
	dir = filepath.Join(dir, "ts")
	bob, err := open(ctx, filepath.Join(dir, "bob"))
	if err != nil {
		return nil, nil, err
	}
	mike, err := open(ctx, filepath.Join(dir, "mike"))
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, bobKey{}, bob)
	ctx = context.WithValue(ctx, mikeKey{}, mike)
	return ctx, resourceList{bob, mike}, nil
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

type bobKey struct{}

type mikeKey struct{}

func GetBob(ctx context.Context) *badger.DB {
	return ctx.Value(bobKey{}).(*badger.DB)
}

func GetMike(ctx context.Context) *badger.DB {
	return ctx.Value(mikeKey{}).(*badger.DB)
}
