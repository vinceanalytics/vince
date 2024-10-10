package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"log"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/vinceanalytics/vince/fb"
	"github.com/vinceanalytics/vince/internal/roaring"
)

func main() {
	iso01()
	iso02()
	writeIso()
	geoname()
	writeGeoname()
}

var (
	xxCountry [4][]string
	xxRegion  [2][]string
)

func writeIso() {
	s := make([]string, len(_iso_1))

	for i := range _iso_1 {
		s[i] = _iso_1[i].A3
	}
	xxCountry[0] = slices.Clone(s)
	for i := range _iso_1 {
		s[i] = _iso_1[i].A
	}
	xxCountry[1] = slices.Clone(s)
	for i := range _iso_1 {
		s[i] = _iso_1[i].Flag
	}
	xxCountry[2] = slices.Clone(s)
	for i := range _iso_1 {
		s[i] = _iso_1[i].Name
	}
	xxCountry[3] = slices.Clone(s)

	s = s[:0]
	for i := range _iso_2 {
		s = append(s, _iso_2[i].Code)
	}
	xxRegion[0] = slices.Clone(s)
	s = s[:0]
	for i := range _iso_2 {
		s = append(s, _iso_2[i].Name)
	}
	xxRegion[1] = slices.Clone(s)
}

type Country struct {
	A    string `json:"alpha_2"`
	A3   string `json:"alpha_3"`
	Flag string `json:"flag"`
	Name string `json:"name"`
}

var (
	_iso_1 []*Country
	_iso_2 []*Region
	code   = map[string]int{}
)

type Region struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func iso02() {
	f, err := os.Open("iso_3166-2.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	type D struct {
		V []*Region `json:"3166-2"`
	}
	o := &D{}
	err = json.NewDecoder(f).Decode(o)
	if err != nil {
		log.Fatal(err)
	}
	_iso_2 = o.V
	slices.SortFunc(_iso_2, func(a, b *Region) int {
		return strings.Compare(a.Code, b.Code)
	})
}

func iso01() {
	f, err := os.Open("iso_3166-1.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	type D struct {
		V []*Country `json:"3166-1"`
	}
	o := &D{}
	err = json.NewDecoder(f).Decode(o)
	if err != nil {
		log.Fatal(err)
	}
	_iso_1 = o.V
	slices.SortFunc(_iso_1, func(a, b *Country) int {
		return strings.Compare(a.A, b.A)
	})
	for i := range _iso_1 {
		code[_iso_1[i].A] = i
	}
}

var (
	_geo_name geoMapping
	city      = roaring.NewDefaultBSI()
	cityCode  = roaring.NewDefaultBSI()
)

type geoMapping struct {
	names []string
	codes []int
}

func (g *geoMapping) Len() int {
	return len(g.names)
}

func (g *geoMapping) Less(i, j int) bool {
	return g.names[i] < g.names[j]
}

func (g *geoMapping) Swap(i, j int) {
	g.names[i], g.names[j] = g.names[j], g.names[i]
	g.codes[i], g.codes[j] = g.codes[j], g.codes[i]
}

func writeGeoname() {
	sort.Sort(&_geo_name)
	for i := range _geo_name.names {
		city.SetValue(uint64(_geo_name.codes[i]), int64(i))
		cityCode.SetValue(uint64(i), int64(_geo_name.codes[i]))
	}
	sx := fb.Serialize(_geo_name.names, xxCountry, xxRegion, city.ToBuffer(), cityCode.ToBuffer())
	os.WriteFile("location.fbs.bin", sx, 0600)
}

func geoname() {
	r, err := zip.OpenReader("allCountries.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	o, err := r.File[0].Open()
	if err != nil {
		log.Fatal(o)
	}
	defer o.Close()

	s := bufio.NewScanner(o)
	sParse(s, 0, func(raw []string) bool {
		if len(raw) != 19 {
			return true
		}
		id := raw[0]
		name := raw[1]
		class := raw[6]
		countryCode := raw[8]
		if class == "P" {
			_, ok := code[countryCode]
			if ok {
				iv, err := strconv.Atoi(id)
				if err != nil {
					log.Fatal("parsing id", id, err)
				}
				_geo_name.names = append(_geo_name.names, countryCode+"-"+name)
				_geo_name.codes = append(_geo_name.codes, iv)
			}
		}
		return true
	})
}

const commentSymbol = byte('#')

func sParse(s *bufio.Scanner, headerLength uint, f func([]string) bool) {
	var err error
	var line string
	var rawSplit []string
	for s.Scan() {
		if headerLength != 0 {
			headerLength--
			continue
		}
		line = s.Text()
		if len(line) == 0 {
			continue
		}
		if line[0] == commentSymbol {
			continue
		}
		rawSplit = strings.Split(line, "\t")
		if !f(rawSplit) {
			break
		}
	}
	if err = s.Err(); err != nil {
		log.Fatal(err)
	}
}
