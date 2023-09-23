package engine

import (
	"context"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/db"
)

type Engine struct {
	*sqle.Engine
}

func New(ctx context.Context) *Engine {
	pro := &Provider{
		db:     db.Get(ctx),
		reader: b3.GetReader(ctx),
	}
	e := sqle.NewDefault(pro)
	setupAuth(ctx, e)
	e.ReadOnly.Store(true)
	return &Engine{Engine: e}
}

type engineKey struct{}

func Open(ctx context.Context) (context.Context, *Engine) {
	e := New(ctx)
	return context.WithValue(ctx, engineKey{}, e), e
}

func Get(ctx context.Context) *Engine {
	return ctx.Value(engineKey{}).(*Engine)
}
