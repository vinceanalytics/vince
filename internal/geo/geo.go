package geo

import (
	"context"
	"net"

	"github.com/oschwald/maxminddb-golang"
	"github.com/vinceanalytics/vince/internal/logger"
)

type Geo struct {
	db *maxminddb.Reader
}

func Open(path string) *Geo {
	if path == "" {
		return &Geo{}
	}
	db, err := maxminddb.Open(path)
	if err != nil {
		logger.Fail("failed opening geoip database", "err", err)
	}
	return &Geo{db: db}
}

func (g *Geo) Get(ip net.IP) Info {
	if g.db == nil {
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
	if g.db != nil {
		return g.db.Close()
	}
	return nil
}

type Record struct {
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

type geoKey struct{}

func With(ctx context.Context, g *Geo) context.Context {
	return context.WithValue(ctx, geoKey{}, g)
}

func Get(ctx context.Context) *Geo {
	return ctx.Value(geoKey{}).(*Geo)
}
