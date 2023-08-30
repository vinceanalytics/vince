package query

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/go-sql-driver/mysql"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func DSN(addr string, a *v1.Client_Auth, tls bool) string {
	x := mysql.Config{
		User:                    a.Name,
		Passwd:                  a.Token,
		Net:                     "tcp",
		Addr:                    addr,
		DBName:                  "vince",
		AllowNativePasswords:    true,
		AllowCleartextPasswords: true,
		Params: map[string]string{
			"tls": strconv.FormatBool(tls),
		},
	}
	return x.FormatDSN()
}

func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}

// TTL is the duration to keep internal database connection
const TTL = time.Hour

// We use same api to access mysql like other clients. This is for internal use
// only.
var clients = must.Must(ristretto.NewCache(&ristretto.Config{
	NumCounters: 10 * 64,
	MaxCost:     64,
	BufferItems: 64,
	OnEvict: func(item *ristretto.Item) {
		item.Value.(*sql.DB).Close()
	},
}))("failed creating clients cache")

func GetInternalClient(ctx context.Context) *sql.DB {
	a := core.GetAuth(ctx)
	if x, ok := clients.Get(a.Name); ok {
		return x.(*sql.DB)
	}
	o := config.Get(ctx)
	dns := DSN(o.MysqlListenAddress, a, config.IsTLS(o))
	db := must.Must(Open(dns))(
		"failed to open mysql db connection for internal client",
	)
	db.SetConnMaxLifetime(time.Minute)
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(5)
	clients.SetWithTTL(a.Name, db, 1, TTL)
	return db
}
