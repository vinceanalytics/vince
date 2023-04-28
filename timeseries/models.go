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
	"github.com/gernest/vince/pkg/log"
	"github.com/klauspost/compress/zstd"
)

func Open(ctx context.Context, dataPath string) (context.Context, io.Closer, error) {
	dir := filepath.Join(dataPath, "ts")
	mike, err := open(ctx, filepath.Join(dir, "mike"))
	if err != nil {
		return nil, nil, err
	}
	geo, err := openGeo(ctx, filepath.Join(dir, "geo"))
	if err != nil {
		mike.Close()
		return nil, nil, err
	}
	sys, err := openSystem(dataPath)
	if err != nil {
		mike.Close()
		geo.Close()
		return nil, nil, err
	}
	ctx = context.WithValue(ctx, mikeKey{}, mike)
	ctx = context.WithValue(ctx, geoKey{}, geo)
	ctx = context.WithValue(ctx, systemKey{}, sys)
	ctx = SetMap(ctx, NewMap())
	return ctx, resourceList{mike, geo, sys}, nil
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

type mikeKey struct{}

type geoKey struct{}

type systemKey struct{}

func GetMike(ctx context.Context) *badger.DB {
	return ctx.Value(mikeKey{}).(*badger.DB)
}

func GetGeo(ctx context.Context) *badger.DB {
	return ctx.Value(geoKey{}).(*badger.DB)
}

func GetSystem(ctx context.Context) *AllSystem {
	return ctx.Value(systemKey{}).(*AllSystem)
}
