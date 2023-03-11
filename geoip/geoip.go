package geoip

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net"
	"sync"

	"github.com/klauspost/compress/zstd"
	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
)

//go:generate go run download/make_mmdb.go

//go:embed dbip-country.mmdb
var data []byte

var (
	mmdb *geoip2.Reader
	once sync.Once
)

func Get() *geoip2.Reader {
	once.Do(func() {
		var err error
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			panic("failed to read embedded mmdb data file gzip data expected " + err.Error())
		}
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
