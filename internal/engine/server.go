package engine

import (
	"bytes"
	"context"
	"crypto/tls"

	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
)

func Listen(ctx context.Context) (*Server, error) {
	o := config.Get(ctx)
	e := Get(ctx)

	// load sites tables
	db.Get(ctx).View(func(txn *badger.Txn) error {
		sites := []byte(keys.Site("") + "/")
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = false
		o.Prefix = sites
		it := txn.NewIterator(o)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			name := bytes.TrimPrefix(it.Item().Key(), sites)
			e.Add(string(name))
		}
		return nil
	})

	sm := server.NewSessionManager(server.DefaultSessionBuilder,
		sql.NoopTracer, e.Analyzer.Catalog.Database, e.MemoryManager,
		e.ProcessList, o.MysqlListenAddress)
	handler := &Handler{
		e:  e.Engine,
		sm: sm,
	}
	l, err := server.NewListener("tcp", o.MysqlListenAddress, "")
	if err != nil {
		return nil, err
	}
	listenerCfg := mysql.ListenerConfig{
		Listener:                 l,
		AuthServer:               &Auth{DB: db.Get(ctx)},
		Handler:                  handler,
		ConnReadBufferSize:       mysql.DefaultConnBufferSize,
		AllowClearTextWithoutTLS: true,
	}
	ls, err := mysql.NewListenerWithConfig(listenerCfg)
	if err != nil {
		return nil, err
	}
	if config.IsTLS(o) {
		cert, err := tls.LoadX509KeyPair(o.TlsCertFile, o.TlsKeyFile)
		if err != nil {
			return nil, err
		}
		ls.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		ls.RequireSecureTransport = true
	}
	return &Server{Listener: ls}, nil
}

type Server struct {
	*mysql.Listener
}

func (s *Server) Start() error {
	s.Accept()
	return nil
}

func (s *Server) Close() error {
	s.Listener.Close()
	return nil
}
