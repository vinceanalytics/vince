package location

import (
	_ "embed"
	"slices"
	"strings"
)

//go:embed city.protobuf.gz
var cityData []byte

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Flag string `json:"flag"`
}

func GetCountry(code string) *Country {
	i, ok := slices.BinarySearch(_iso_1_code, code)
	if !ok {
		return &Country{Code: code}
	}
	return &Country{
		Code: code,
		Name: _iso_1_name[i],
		Flag: _iso_1_flag[i],
	}
}

type Region struct {
	Name string `json:"name"`
	Flag string `json:"flag"`
}

func GetRegion(code string) *Region {
	i, ok := slices.BinarySearch(_iso_2_code, code)
	if !ok {
		return &Region{Name: code}
	}
	name := _iso_2_name[i]
	var flag string
	a, _, _ := strings.Cut(name, "-")
	if n, ok := slices.BinarySearch(_iso_1_name, a); ok {
		flag = _iso_1_flag[n]
	}
	return &Region{Name: name, Flag: flag}
}
