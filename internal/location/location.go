package location

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
	"slices"
	"strings"
	"sync"

	"github.com/RoaringBitmap/roaring/v2/roaring64"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"google.golang.org/protobuf/proto"
)

//go:generate go run gen/main.go

//go:embed city.protobuf.gz
var cityData []byte

var (
	_geo_name []string
	city      = roaring64.NewDefaultBSI()
	cityCode  = roaring64.NewDefaultBSI()
)
var once sync.Once

type City struct {
	Name string `json:"name"`
	Flag string `json:"country_flag"`
}

func GetCity(code uint32) City {
	once.Do(func() {
		r, _ := gzip.NewReader(bytes.NewReader(cityData))
		all, _ := io.ReadAll(r)
		var l v1.Location
		proto.Unmarshal(all, &l)
		_geo_name = l.Names
		city.ReadFrom(bytes.NewReader(l.City))
		cityCode.ReadFrom(bytes.NewReader(l.CityCode))
		l.Reset()
	})
	v, ok := city.GetValue(uint64(code))
	if !ok {
		return City{Name: "N/A"}
	}
	idx, _ := cityCode.GetValue(uint64(v))
	return City{
		Name: _geo_name[v],
		Flag: _iso_1_flag[idx],
	}
}

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

func GetCountry(code string) Country {
	i, ok := slices.BinarySearch(_iso_1_code, code)
	if !ok {
		return Country{Code: code}
	}
	return Country{
		Code: code,
		Name: _iso_1_name[i],
		Flag: _iso_1_flag[i],
	}
}

type Region struct {
	Name string `json:"name"`
	Flag string `json:"flag"`
}

func GetRegion(code string) Region {
	i, ok := slices.BinarySearch(_iso_2_code, code)
	if !ok {
		return Region{Name: code}
	}
	name := _iso_2_name[i]
	var flag string
	a, _, _ := strings.Cut(code, "-")
	if n, ok := slices.BinarySearch(_iso_1_code, a); ok {
		flag = _iso_1_flag[n]
	}
	return Region{Name: name, Flag: flag}
}
