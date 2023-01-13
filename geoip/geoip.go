package geoip

import (
	"bytes"
	_ "embed"
	"io"
	"net"
	"sync"

	"github.com/klauspost/compress/zstd"
	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

// compressed with the command
//
//	zstd -o=geoip/dbip-country.mmdb.zstd -22 --ultra  geoip/dbip-country.mmdb
//
//go:embed dbip-country.mmdb.zstd
var data []byte

var (
	mmdb *geoip2.Reader
	once sync.Once
)

func Get() *geoip2.Reader {
	once.Do(func() {
		r, err := zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			panic(err.Error())
		}
		defer r.Close()
		b, err := io.ReadAll(r)
		if err != nil {
			panic(err.Error())
		}
		mmdb, err = geoip2.FromBytes(b)
		if err != nil {
			panic(err.Error())
		}
	})
	return mmdb
}

func Reader() *maxminddb.Reader {
	r, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		panic(err.Error())
	}
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err.Error())
	}
	reader, err := maxminddb.FromBytes(b)
	if err != nil {
		panic(err.Error())
	}
	return reader
}

func Lookup(ip net.IP) (*geoip2.City, error) {
	return Get().City(ip)
}
