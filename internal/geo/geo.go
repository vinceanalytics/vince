package geo

import (
	_ "embed"
	"fmt"
	"net"
	"slices"
	"sync"

	"github.com/oschwald/maxminddb-golang"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/models"
)

//go:embed city.mmdb
var city []byte

var (
	reader, _ = maxminddb.FromBytes(city)
)

//go:generate go run gen/main.go

var cityPool = &sync.Pool{New: func() any { return new(City) }}

type City struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Country struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
	Subdivisions []struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
}

func UpdateCity(ip net.IP, m *models.Model) error {
	x := cityPool.Get().(*City)
	err := reader.Lookup(ip, x)
	if err != nil {
		return err
	}
	if x.Country.IsoCode == "" {
		return nil
	}
	m.Country = []byte(x.Country.IsoCode)
	if len(x.City.Names) > 0 && x.City.Names["en"] != "" {
		m.City = location.GetCityCode(x.Country.IsoCode, x.City.Names["en"])
	}
	if len(x.Subdivisions) > 0 {
		m.Subdivision1Code = location.GetRegionCode(x.Subdivisions[0].Names["en"])
	}
	*x = City{}
	cityPool.Put(x)
	return nil
}

func Rand(size int) []string {
	n := reader.Networks(maxminddb.SkipAliasedNetworks)
	var a City
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
