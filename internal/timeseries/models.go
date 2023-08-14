package timeseries

import (
	"context"
	"io"

	"github.com/vinceanalytics/vince/internal/neo"
)

func Open(ctx context.Context, dir string) (context.Context, io.Closer) {
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
