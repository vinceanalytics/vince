package geo

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

//go:embed country.gz
var country []byte

var (
	mmdb *geoip2.Reader
	once sync.Once
)

func Get(ip net.IP) (Info, error) {
	x, err := get().City(ip)
	if err != nil {
		return Info{}, err
	}
	var region string
	if len(x.Subdivisions) > 0 {
		region = x.Subdivisions[0].Names["en"]
	}
	return Info{
		City:    x.City.Names["en"],
		Country: x.Country.Names["en"],
		Region:  region,
	}, nil
}

type Info struct {
	City    string
	Country string
	Region  string
}

func get() *geoip2.Reader {
	once.Do(func() {
		var err error
		r, err := gzip.NewReader(bytes.NewReader(country))
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
