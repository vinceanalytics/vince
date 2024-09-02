package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go/format"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
	"google.golang.org/protobuf/proto"
)

func main() {
	iso01()
	iso02()
	writeIso()
	geoname()
	writeGeoname()
}

func writeIso() {
	b := new(bytes.Buffer)
	fmt.Fprintln(b, "package location")
	s := make([]string, len(_iso_1))
	for i := range _iso_1 {
		s[i] = _iso_1[i].A
	}
	fmt.Fprintln(b, "var (")
	for i := range _iso_1 {
		s[i] = _iso_1[i].A
	}
	fmt.Fprintf(b, "\n _iso_1_code =%#v", s)
	for i := range _iso_1 {
		s[i] = _iso_1[i].Flag
	}
	fmt.Fprintf(b, "\n _iso_1_flag =%#v", s)
	for i := range _iso_1 {
		s[i] = _iso_1[i].Name
	}
	fmt.Fprintf(b, "\n _iso_1_name =%#v", s)

	s = s[:0]
	for i := range _iso_2 {
		s = append(s, _iso_2[i].Code)
	}
	fmt.Fprintf(b, "\n _iso_2_code =%#v", s)
	s = s[:0]
	for i := range _iso_2 {
		s = append(s, _iso_2[i].Name)
	}
	fmt.Fprintf(b, "\n _iso_2_name =%#v", s)
	fmt.Fprintln(b, ")")

	data, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("iso.go", data, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

type Country struct {
	A    string `json:"alpha_2"`
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
		return strings.Compare(a.Name, b.Name)
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
	_geo_name []string
	city      = roaring64.NewDefaultBSI()
	cityCode  = roaring64.NewDefaultBSI()
)

func writeGeoname() {
	loc := &v1.Location{
		Names: _geo_name,
	}
	b := new(bytes.Buffer)
	cityCode.RunOptimize()
	b.Reset()
	_, err := cityCode.WriteTo(b)
	if err != nil {
		log.Fatal(err)
	}
	loc.CityCode = bytes.Clone(b.Bytes())

	city.RunOptimize()
	b.Reset()
	_, err = city.WriteTo(b)
	if err != nil {
		log.Fatal(err)
	}
	loc.City = b.Bytes()

	data, _ := proto.Marshal(loc)
	err = os.WriteFile("city.protobuf.gz", compress(data), 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func compress(data []byte) []byte {
	b := new(bytes.Buffer)
	w := gzip.NewWriter(b)
	w.Write(data)
	w.Close()
	return b.Bytes()
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
			c, ok := code[countryCode]
			if ok {
				iv, err := strconv.Atoi(id)
				if err != nil {
					log.Fatal("parsing id", id, err)
				}
				column := uint64(iv)
				idx := len(_geo_name)
				_geo_name = append(_geo_name, name)
				city.SetValue(column, int64(idx))
				cityCode.SetValue(uint64(idx), int64(c))
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
