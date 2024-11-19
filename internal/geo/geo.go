package geo

import (
	"embed"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/oschwald/maxminddb-golang"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/models"
)

//go:embed data
var all embed.FS

//go:generate go run gen/main.go chunk city.mmdb

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

type Geo struct {
	lo *location.Location
	db *maxminddb.Reader
}

func (g *Geo) Close() error {
	return g.db.Close()
}

func New(lo *location.Location, dir string) (*Geo, error) {
	path := filepath.Join(dir, "city.mmdb")
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		slog.Info("unpacking geo database", "path", path)
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		ls, _ := all.ReadDir("data")
		for _, n := range ls {
			fn, _ := all.Open(filepath.Join("data", n.Name()))
			f.ReadFrom(fn)
		}
		f.Sync()
		stat, _ := f.Stat()
		err = f.Close()
		if err != nil {
			return nil, err
		}
		slog.Info("unpacked geo database", "size_in_mb", stat.Size()/(1<<20))
	}
	r, err := maxminddb.Open(path)
	if err != nil {
		return nil, err
	}
	return &Geo{db: r, lo: lo}, nil
}

func (g *Geo) UpdateCity(ip net.IP, m *models.Model) error {
	x := cityPool.Get().(*City)
	err := g.db.Lookup(ip, x)
	if err != nil {
		return err
	}
	if x.Country.IsoCode == "" {
		return nil
	}
	m.Country = []byte(x.Country.IsoCode)
	if len(x.City.Names) > 0 && x.City.Names["en"] != "" {
		m.City = g.lo.GetCityCode(x.Country.IsoCode, x.City.Names["en"])
	}
	if len(x.Subdivisions) > 0 {
		m.Subdivision1Code = g.lo.GetRegionCode(x.Subdivisions[0].Names["en"])
	}
	*x = City{}
	cityPool.Put(x)
	return nil
}

func (g *Geo) Rand(size int) []string {
	n := g.db.Networks(maxminddb.SkipAliasedNetworks)
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
