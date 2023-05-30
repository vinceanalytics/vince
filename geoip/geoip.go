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

//go:embed city.mmdb
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

type Info struct {
	City    string
	Country string
	Region  string
}

func Lookup(ip net.IP) Info {
	x, err := Get().City(ip)
	if err != nil {
		// log error
		return Info{}
	}
	return Info{
		City:    x.City.Names["en"],
		Country: x.Country.Names["en"],
		Region:  x.Subdivisions[0].Names["en"],
	}
}
