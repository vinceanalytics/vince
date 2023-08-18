package geoip

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
	"github.com/vinceanalytics/vince/internal/must"
)

//go:generate go run download/make_mmdb.go

//go:embed city.mmdb
var CityData []byte

var (
	mmdb *geoip2.Reader
	once sync.Once
)

type Info struct {
	City    string
	Country string
	Region  string
}

func Lookup(ip net.IP) Info {
	x, err := get().City(ip)
	if err != nil {
		// log error
		return Info{}
	}
	var region string
	if len(x.Subdivisions) > 0 {
		region = x.Subdivisions[0].Names["en"]
	}
	return Info{
		City:    x.City.Names["en"],
		Country: x.Country.Names["en"],
		Region:  region,
	}
}

func get() *geoip2.Reader {
	once.Do(func() {
		r := must.Must(gzip.NewReader(bytes.NewReader(CityData)))(
			"failed to read embedded mmdb data file gzip data expected ",
		)
		b := must.Must(io.ReadAll(r))("failed reading compressed mmdb")
		mmdb = must.Must(geoip2.FromBytes(b))(
			"failed opening mmdb file",
		)
	})
	return mmdb
}
