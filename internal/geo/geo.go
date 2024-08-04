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
	o := Info{
		CountryCode:   x.Country.IsoCode,
		CityGeonameID: uint32(x.City.GeoNameID),
	}

	if o.CountryCode != "" && len(x.Subdivisions) > 0 {
		o.SubDivision1Code = o.CountryCode + "-" + x.Subdivisions[0].IsoCode
	}
	if o.CountryCode != "" && len(x.Subdivisions) > 1 {
		o.SubDivision2Code = o.CountryCode + "-" + x.Subdivisions[1].IsoCode
	}
	return o, nil
}

type Info struct {
	CountryCode      string
	SubDivision1Code string
	SubDivision2Code string
	CityGeonameID    uint32
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
