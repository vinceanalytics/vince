package engine

import (
	"context"

	sqle "github.com/dolthub/go-mysql-server"
)

type Engine struct {
	DB *RadOnly
	*sqle.Engine
}

func New() *Engine {
	db := Database("vince")
	return &Engine{
		Engine: sqle.NewDefault(&Provider{
			base: db,
		}),
		DB: db,
	}
}

type engineKey struct{}

func Open(ctx context.Context) (context.Context, *Engine) {
	e := New()
	return context.WithValue(ctx, engineKey{}, e), e
}

func Get(ctx context.Context) *Engine {
	return ctx.Value(engineKey{}).(*Engine)
}
