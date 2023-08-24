package engine

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/vinceanalytics/vince/internal/db"
)

type Engine struct {
	*sqle.Engine
}

func New(db *badger.DB) *Engine {
	return &Engine{
		Engine: sqle.NewDefault(&Provider{
			db: db,
		}),
	}
}

type engineKey struct{}

func Open(ctx context.Context) (context.Context, *Engine) {
	e := New(db.Get(ctx))
	return context.WithValue(ctx, engineKey{}, e), e
}

func Get(ctx context.Context) *Engine {
	return ctx.Value(engineKey{}).(*Engine)
}
