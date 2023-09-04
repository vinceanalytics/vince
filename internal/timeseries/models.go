package timeseries

import (
	"context"
	"io"

	"github.com/thanos-io/objstore"
	"github.com/vinceanalytics/vince/internal/neo"
)

func Open(ctx context.Context, o objstore.Bucket, capacity int) (context.Context, io.Closer) {
	a := neo.NewIngest(ctx, o, capacity)
	return context.WithValue(ctx, blockKey{}, a), a
}

type blockKey struct{}

func Block(ctx context.Context) *neo.Ingest {
	return ctx.Value(blockKey{}).(*neo.Ingest)
}
