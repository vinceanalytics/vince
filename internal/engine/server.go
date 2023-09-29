package engine

import (
	"context"
	"crypto/tls"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/engine/handler"
)

type Server struct {
	Listener *mysql.Listener
}

func (s *Server) Close() error {
	s.Listener.Close()
	return nil
}

func (s *Server) Shutdown(_ context.Context) error {
	s.Listener.Shutdown()
	return nil
}

func (s *Server) Start() error {
	s.Listener.Accept()
	return nil
}

func Listen(ctx context.Context) (*Server, error) {
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
	h := handler.NewServer(svrConfig, e.Engine, buildSession(ctx), nil)
	ls, err := server.NewListener(svrConfig.Protocol, svrConfig.Address, svrConfig.Socket)
	if err != nil {
		return nil, err
	}

	listenerCfg := mysql.ListenerConfig{
		Listener:                 ls,
		AuthServer:               e.Analyzer.Catalog.MySQLDb,
		Handler:                  h,
		ConnReadTimeout:          svrConfig.ConnReadTimeout,
		ConnWriteTimeout:         svrConfig.ConnWriteTimeout,
		MaxConns:                 svrConfig.MaxConnections,
		ConnReadBufferSize:       mysql.DefaultConnBufferSize,
		AllowClearTextWithoutTLS: svrConfig.AllowClearTextWithoutTLS,
	}
	vtListnr, err := mysql.NewListenerWithConfig(listenerCfg)
	if err != nil {
		return nil, err
	}

	if svrConfig.Version != "" {
		vtListnr.ServerVersion = svrConfig.Version
	}
	vtListnr.TLSConfig = svrConfig.TLSConfig
	vtListnr.RequireSecureTransport = svrConfig.RequireSecureTransport

	return &Server{Listener: vtListnr}, nil
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
