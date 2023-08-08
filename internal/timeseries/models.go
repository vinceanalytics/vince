package timeseries

import (
	"context"
	"io"
	"path/filepath"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/neo"
)

func Open(ctx context.Context, o *config.Options) (context.Context, io.Closer) {
	dir := filepath.Join(o.DataPath, "ts")
	a := neo.NewBlock(dir)
	return context.WithValue(ctx, blockKey{}, a), a
}

type blockKey struct{}

func Block(ctx context.Context) *neo.ActiveBlock {
	return ctx.Value(blockKey{}).(*neo.ActiveBlock)
}

func Save(ctx context.Context) {
	Block(ctx).Save(ctx)
}
