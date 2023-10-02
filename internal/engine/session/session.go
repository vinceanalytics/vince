package session

import (
	"context"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/engine/auth"
	"github.com/vinceanalytics/vince/internal/scopes"
)

type Session struct {
	sql.Session
	Claim *auth.Claim
	ctx   context.Context
}

func New(ctx context.Context, sess sql.Session, claim *auth.Claim) *Session {
	return &Session{
		Session: sess,
		Claim:   claim,
		ctx:     ctx,
	}
}

func (s *Session) Context() context.Context {
	return s.ctx
}

func (s *Session) DB() db.Provider {
	return db.Get(s.ctx)
}

func (s *Session) B3() b3.Reader {
	return b3.GetReader(s.ctx)
}

func Get(ctx *sql.Context) *Session {
	return ctx.Session.(*Session)
}

func access(scope scopes.Scope) error {
	return mysql.NewSQLError(mysql.ERAccessDeniedError, mysql.SSAccessDeniedError,
		"Access denied for operation '%v'", scope.String())
}

func (s *Session) Allow(action scopes.Scope) error {
	if s.Claim == nil || !s.Claim.Claims.ValidFor(action) {
		return access(action)
	}
	return nil
}
