package engine

import (
	"context"

	"github.com/apache/arrow/go/v14/parquet"
	"github.com/dgraph-io/badger/v4"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/timeseries"
)

type Engine struct {
	*sqle.Engine
}

func New(ctx Context) *Engine {
	return &Engine{
		Engine: sqle.NewDefault(&Provider{
			Context: ctx,
		}),
	}
}

type engineKey struct{}

func Open(ctx context.Context) (context.Context, *Engine) {
	e := New(Context{
		DB:                   db.Get(ctx),
		ReadBlock:            timeseries.Block(ctx).ReadBlock,
		PrepareTableForQuery: timeseries.Block(ctx).Commit,
	})
	return context.WithValue(ctx, engineKey{}, e), e
}

func Get(ctx context.Context) *Engine {
	return ctx.Value(engineKey{}).(*Engine)
}

type Context struct {
	DB                   *badger.DB
	ReadBlock            func(ulid.ULID, func(parquet.ReaderAtSeeker))
	PrepareTableForQuery func(string)
}
