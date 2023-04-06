package timeseries

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
	"github.com/gernest/vince/cities"
	"github.com/gernest/vince/log"
	"github.com/klauspost/compress/zstd"
)

// Creates two badger.DB instances one for temporary aggregate and another for
// permanent storage/query.
func Open(ctx context.Context, dir string) (context.Context, io.Closer, error) {
	dir = filepath.Join(dir, "ts")
	bob, err := openBob(ctx, filepath.Join(dir, "bob"))
	if err != nil {
		return nil, nil, err
	}
	mike, err := open(ctx, filepath.Join(dir, "mike"))
	if err != nil {
		return nil, nil, err
	}
	geo, err := openGeo(ctx, filepath.Join(dir, "geo"))
	if err != nil {
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, bobKey{}, bob)
	ctx = context.WithValue(ctx, mikeKey{}, mike)
	ctx = context.WithValue(ctx, geoKey{}, geo)
	return ctx, resourceList{bob, mike, geo}, nil
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

func openBob(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD).
		// keep up to two versions.
		WithNumVersionsToKeep(2)
	return badger.Open(o)
}

func openGeo(ctx context.Context, path string) (*badger.DB, error) {
	os.MkdirAll(path, 0755)
	o := badger.DefaultOptions(path).
		WithLogger(log.Badger(ctx)).
		WithCompression(options.ZSTD)
	db, err := badger.Open(o)
	if err != nil {
		return nil, err
	}
	size, _ := db.Size()
	if size == 0 {
		lx := log.Get(ctx)
		lx.Info().Msg("building geoname index, its going to take a while ...")
		// restore the index
		d, err := zstd.NewReader(bytes.NewReader(cities.GeonameIndex))
		if err != nil {
			return nil, err
		}
		err = db.Load(d, runtime.NumCPU())
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

type bobKey struct{}

type mikeKey struct{}

type geoKey struct{}

func GetBob(ctx context.Context) *badger.DB {
	return ctx.Value(bobKey{}).(*badger.DB)
}

func GetMike(ctx context.Context) *badger.DB {
	return ctx.Value(mikeKey{}).(*badger.DB)
}

func GetGeo(ctx context.Context) *badger.DB {
	return ctx.Value(geoKey{}).(*badger.DB)
}
