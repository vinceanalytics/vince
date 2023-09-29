package engine

import (
	"context"
	"crypto/tls"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/vinceanalytics/vince/internal/config"
)

func Listen(ctx context.Context) (*server.Server, error) {
	o := config.Get(ctx)
	e := Get(ctx)
	svrConfig := server.Config{
		Protocol:                     "tcp",
		Address:                      o.MysqlListenAddress,
		Socket:                       config.SocketFile(o),
		DisableClientMultiStatements: true,
	}
	if config.IsTLS(o) {
		cert, err := tls.LoadX509KeyPair(o.TlsCertFile, o.TlsKeyFile)
		if err != nil {
			return nil, err
		}
		svrConfig.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		svrConfig.RequireSecureTransport = true
	} else {
		svrConfig.AllowClearTextWithoutTLS = true
	}
	return server.NewServer(svrConfig, e.Engine, buildSession(ctx), nil)
}

func buildSession(base context.Context) server.SessionBuilder {
	return func(ctx context.Context, conn *mysql.Conn, addr string) (sql.Session, error) {
		s, err := server.DefaultSessionBuilder(ctx, conn, addr)
		if err != nil {
			return nil, err
		}
		return &Session{Session: s, base: func() context.Context { return base }}, nil
	}
}
