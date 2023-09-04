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
		DB:     db.Get(ctx),
		Reader: b3.GetReader(ctx),
	})
	return context.WithValue(ctx, engineKey{}, e), e
}

func Get(ctx context.Context) *Engine {
	return ctx.Value(engineKey{}).(*Engine)
}

type Context struct {
	DB     db.Provider
	Reader b3.Reader
}
