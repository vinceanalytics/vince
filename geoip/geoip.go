package geoip

import (
	"bytes"
	_ "embed"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
	"github.com/oschwald/geoip2-golang"
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
