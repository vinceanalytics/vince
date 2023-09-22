package engine

import (
	"context"
	"crypto/tls"

	"github.com/dolthub/go-mysql-server/server"
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
	return server.NewDefaultServer(svrConfig, e.Engine)
}
