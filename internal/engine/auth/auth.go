package auth

import (
	"context"
	"net"

	"github.com/dolthub/vitess/go/mysql"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/vinceanalytics/vince/internal/scopes"
	"github.com/vinceanalytics/vince/internal/secrets"
	"github.com/vinceanalytics/vince/internal/tokens"
)

type Auth struct {
	ctx context.Context
}

func New(ctx context.Context) *Auth {
	return &Auth{ctx: ctx}
}

var _ mysql.AuthServer = (*Auth)(nil)

func (a *Auth) Salt() ([]byte, error) {
	return mysql.NewSalt()
}

func (a *Auth) ValidateHash(salt []byte, user string, authResponse []byte, remoteAddr net.Addr) (mysql.Getter, error) {
	return &mysql.StaticUserData{}, mysql.NewSQLError(mysql.ERAccessDeniedError, mysql.SSAccessDeniedError, "Access denied for user '%v'", user)
}

func (a *Auth) AuthMethod(user, addr string) (string, error) {
	return mysql.MysqlClearPassword, nil
}

func (a *Auth) Negotiate(c *mysql.Conn, user string, remoteAddr net.Addr) (mysql.Getter, error) {
	// Finish the negotiation.
	password, err := mysql.AuthServerNegotiateClearOrDialog(c, mysql.MysqlClearPassword)
	if err != nil {
		return nil, err
	}
	claim, ok := tokens.ValidWithClaims(
		secrets.Get(a.ctx),
		password, scopes.Query)
	if !ok {
		return &mysql.StaticUserData{}, mysql.NewSQLError(mysql.ERAccessDeniedError, mysql.SSAccessDeniedError, "Access denied for user '%v'", user)
	}
	return &Claim{Username: user, Claims: claim}, nil
}

type Claim struct {
	Username string
	Claims   *tokens.Claims
}

func (c *Claim) Get() *query.VTGateCallerID {
	return &query.VTGateCallerID{Username: c.Username}
}
