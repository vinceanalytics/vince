package engine

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
)

type Session struct {
	sql.Session
	base func() context.Context
}

func (s *Session) Context() context.Context {
	return s.base()
}

func GetSession(ctx *sql.Context) *Session {
	return ctx.Session.(*Session)
}
