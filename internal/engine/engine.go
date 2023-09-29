package engine

import (
	"context"

	sqle "github.com/dolthub/go-mysql-server"
)

type Engine struct {
	*sqle.Engine
}

func New(ctx context.Context) *Engine {
	pro := NewProvider()
	e := sqle.NewDefault(pro)
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
