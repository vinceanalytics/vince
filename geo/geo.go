package geo

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
)

type Geo struct {
	db *maxminddb.Reader
}

func Open(path string) (*Geo, error) {
	if path == "" {
		return nil, nil
	}
	db, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	return &Geo{db: db}, nil
}

func (g *Geo) Get(ip net.IP) Info {
	if g == nil {
		return Info{}
	}
	var x City
	g.db.Lookup(ip, &x)
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

func (g *Geo) Close() error {
	if g == nil {
		return nil
	}
	return g.db.Close()
}

type City struct {
	City struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Subdivisions []struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
	Country struct {
		Names map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
}

type Info struct {
	City    string
	Country string
	Region  string
}
