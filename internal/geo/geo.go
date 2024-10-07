package geo

import (
	_ "embed"
	"fmt"
	"net"
	"slices"

	"github.com/oschwald/geoip2-golang"
	"github.com/oschwald/maxminddb-golang"
	"github.com/vinceanalytics/vince/internal/location"
)

//go:embed city.mmdb
var city []byte

var (
	mmdb, _ = geoip2.FromBytes(city)
)

//go:generate go run gen/main.go

func Get(ip net.IP) (Info, error) {
	x, err := mmdb.City(ip)
	if err != nil {
		return Info{}, err
	}
	o := Info{
		CountryCode:   x.Country.IsoCode,
		CityGeonameID: uint32(x.City.GeoNameID),
	}
	if o.CountryCode != "" && len(x.City.Names) > 0 && x.City.Names["en"] != "" {
		o.CityGeonameID = location.GetCityCode(o.CountryCode, x.City.Names["en"])
	}
	if o.CountryCode != "" && len(x.Subdivisions) > 0 {
		o.SubDivision1Code = location.GetRegionCode(x.Subdivisions[0].Names["en"])
	}
	if o.CountryCode != "" && len(x.Subdivisions) > 1 {
		o.SubDivision2Code = location.GetRegionCode(x.Subdivisions[1].Names["en"])
	}
	return o, nil
}

type Info struct {
	CountryCode      string
	SubDivision1Code []byte
	SubDivision2Code []byte
	CityGeonameID    uint32
}

func Rand(size int) []string {
	reader, err := maxminddb.FromBytes(city)
	if err != nil {
		panic(err.Error())

	}
	n := reader.Networks(maxminddb.SkipAliasedNetworks)
	var a geoip2.City
	m := map[string]struct{}{}
	ips := map[string]struct{}{}
	for n.Next() && len(ips) < size {
		net, err := n.Network(&a)
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, ok := m[a.Country.IsoCode]
		if ok {
			continue
		}
		m[a.Country.IsoCode] = struct{}{}
		ips[net.IP.String()] = struct{}{}
	}
	o := make([]string, 0, size)
	for k := range ips {
		o = append(o, k)
	}
	slices.Sort(o)
	return o
}
