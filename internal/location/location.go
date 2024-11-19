package location

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/cockroachdb/pebble"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/location/lo"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/data"
)

//go:generate flatc --go lo.fbs
//go:generate go run gen/main.go

type Location struct {
	db *pebble.DB
}

func New() *Location {
	db, err := data.Open("data", &pebble.Options{
		ReadOnly:   true,
		DisableWAL: true,
		FS:         &FS{},
	})
	assert.Nil(err)
	return &Location{db: db}
}

func (lo *Location) Close() error {
	return lo.db.Close()
}

func (lo *Location) DB() *pebble.DB {
	return lo.db
}

type City struct {
	Name string `json:"name"`
	Flag string `json:"country_flag"`
}

func (lo *Location) GetReferral(domain string) (code []byte) {
	data.Get(lo.db, append(keys.ReferralPrefix, []byte(domain)...), func(val []byte) error {
		code = bytes.Clone(val)
		return nil
	})
	return
}

func (lo *Location) GetRegionCode(region string) (code []byte) {
	data.Get(lo.db, append(keys.RegionNamePrefix, []byte(region)...), func(val []byte) error {
		code = bytes.Clone(val)
		return nil
	})
	return
}

func (lo *Location) GetCityCode(country, city string) (code uint32) {
	data.Get(lo.db, append(keys.CityNamePrefix, []byte(country+"-"+city)...), func(val []byte) error {
		code = binary.BigEndian.Uint32(val)
		return nil
	})
	return
}

func (l *Location) GetCity(code uint32) (city City) {
	data.Get(l.db, append(keys.CityCodePrefix, binary.BigEndian.AppendUint32(nil, code)...), func(val []byte) error {
		city.Name = string(val)
		return nil
	})
	if city.Name == "" {
		city.Name = "N?A"
		return
	}
	country, name, _ := strings.Cut(city.Name, "-")
	city.Name = name
	data.Get(l.db, append(keys.CountryCodePrefix, []byte(country)...), func(val []byte) error {
		city.Flag = string(lo.GetRootAsCountry(val, 0).Flag())
		return nil
	})
	return
}

type Country struct {
	Alpha string `json:"alpha_3"`
	Code  string `json:"code"`
	Name  string `json:"name"`
	Flag  string `json:"flag"`
}

func (l *Location) GetCountry(code string) (co Country) {
	data.Get(l.db, append(keys.CountryCodePrefix, []byte(code)...), func(val []byte) error {
		v := lo.GetRootAsCountry(val, 0)
		co.Code = code
		co.Alpha = string(v.Alpha3())
		co.Name = string(v.Name())
		co.Flag = string(v.Flag())
		return nil
	})
	return
}

type Region struct {
	Name string `json:"name"`
	Flag string `json:"flag"`
}

var codeSep = []byte("-")

func (l *Location) GetRegion(code []byte) (r Region) {
	data.Get(l.db, append(keys.RegionCodePrefix, code...), func(val []byte) error {
		r.Name = string(val)
		return nil
	})
	if r.Name == "" {
		return
	}
	country, _, _ := bytes.Cut(code, []byte("-"))
	data.Get(l.db, append(keys.CountryCodePrefix, []byte(country)...), func(val []byte) error {
		r.Flag = string(lo.GetRootAsCountry(val, 0).Flag())
		return nil
	})
	return
}
