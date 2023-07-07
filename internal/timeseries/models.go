package timeseries

import (
	"context"
	"errors"
	"io"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/config"
)

func Open(ctx context.Context, o *config.Options) (context.Context, io.Closer, error) {
	dir := filepath.Join(o.DataPath, "ts")
	neo, err := OpenStore(ctx, dir)
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

func Store(ctx context.Context) *V9 {
	return ctx.Value(storeKey{}).(*V9)
}
