package engine

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/db"
)

type Session struct {
	sql.Session
	Claim *Claim
	base  func() context.Context
}

func (s *Session) Context() context.Context {
	return s.base()
}

func (s *Session) DB() db.Provider {
	return db.Get(s.base())
}

func (s *Session) B3() b3.Reader {
	return b3.GetReader(s.base())
}

func GetSession(ctx *sql.Context) *Session {
	return ctx.Session.(*Session)
}
