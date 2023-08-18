package query

import (
	"database/sql"
	"net"
	"net/url"
	"time"

	"github.com/go-sql-driver/mysql"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func ParseDSN(o *v1.Client) string {
	a := o.Instance[o.Active.Instance].Accounts[o.Active.Account]
	u, _ := url.Parse(o.Active.Instance)
	ia, _, _ := net.SplitHostPort(u.Host)
	_, port, _ := net.SplitHostPort(a.Mysql)
	x := mysql.Config{
		User:                 a.Name,
		Passwd:               a.Token,
		Net:                  "tcp",
		Addr:                 ia + ":" + port,
		DBName:               "vince",
		AllowNativePasswords: true,
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
