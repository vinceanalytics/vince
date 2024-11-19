package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/pebble"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/location/lo"
	"github.com/vinceanalytics/vince/internal/util/assert"
	"github.com/vinceanalytics/vince/internal/util/data"
	"gopkg.in/yaml.v2"
)

var ba *pebble.Batch

func main() {
	os.RemoveAll("data")
	db, err := data.Open("data", &pebble.Options{
		DisableWAL: true,
	})
	assert.Nil(err)
	ba = db.NewBatchWithSize(32 << 10)
	iso01()
	iso02()
	writeIso()
	geoname()
	writeGeoname()
	writeReferral()
	assert.Nil(ba.Commit(pebble.NoSync))
	assert.Nil(db.Compact([]byte{0}, []byte{math.MaxUint8}, true))
	assert.Nil(db.Close())
}

var (
	xxCountry [4][]string
	xxRegion  [2][]string
)

func writeIso() {
	key := make([]byte, 0, 1+2)
	key = append(key, keys.CountryCodePrefix...)
	b := flatbuffers.NewBuilder(1 << 10)
	for _, co := range _iso_1 {
		key = append(key[:1], []byte(co.A)...)
		assert.Nil(ba.Set(key, co.Encode(b), pebble.NoSync))
	}
	code := bytes.Clone(keys.RegionCodePrefix)
	name := slices.Clone(keys.RegionNamePrefix)
	for _, co := range _iso_2 {
		code = append(code[:1], []byte(co.Code)...)
		assert.Nil(ba.Set(code, []byte(co.Name), pebble.NoSync))
		name = append(name[:1], []byte(co.Name)...)
		assert.Nil(ba.Set(name, []byte(co.Code), pebble.NoSync))
	}
}

type Country struct {
	A    string `json:"alpha_2"`
	A3   string `json:"alpha_3"`
	Flag string `json:"flag"`
	Name string `json:"name"`
}

func (c *Country) Encode(b *flatbuffers.Builder) []byte {
	b.Reset()
	a2 := b.CreateString(c.A)
	a3 := b.CreateString(c.A3)
	flag := b.CreateString(c.Flag)
	name := b.CreateString(c.Name)

	lo.CountryStart(b)
	lo.CountryAddAlpha2(b, a2)
	lo.CountryAddAlpha3(b, a3)
	lo.CountryAddFlag(b, flag)
	lo.CountryAddName(b, name)
	b.Finish(lo.CountryEnd(b))
	return b.FinishedBytes()
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

func (r *Region) Encode(b *flatbuffers.Builder) []byte {
	b.Reset()
	code := b.CreateString(r.Code)
	name := b.CreateString(r.Name)

	lo.RegionStart(b)
	lo.RegionAddCode(b, code)
	lo.RegionAddName(b, name)
	b.Finish(lo.RegionEnd(b))
	return b.FinishedBytes()
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
	key := make([]byte, 0, 64)
	code := make([]byte, 0, 1+4)
	key = append(key, keys.CityNamePrefix...)
	code = append(code, keys.CityCodePrefix...)
	for i := range _geo_name.names {
		name := _geo_name.names[i]
		co := uint32(_geo_name.codes[i])
		key = append(key[:1], []byte(name)...)
		code = binary.BigEndian.AppendUint32(code[:1], co)
		assert.Nil(ba.Set(key, code[1:], pebble.NoSync))
		assert.Nil(ba.Set(code, key[1:], pebble.NoSync))
	}
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

const srcFileURL = "https://raw.githubusercontent.com/snowplow-referer-parser/referer-parser/master/resources/referers.yml"

type refererData map[string]map[string]map[string][]string

func writeReferral() {
	res, err := http.Get(srcFileURL)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		log.Fatal(res.Status)
	}

	bs, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var data refererData
	err = yaml.NewDecoder(bytes.NewReader(bs)).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}

	key := bytes.Clone(keys.ReferralPrefix)
	for _, m := range data {
		for ref, domains := range m {

			for _, domain := range domains["domains"] {
				key = append(key[:1], []byte(domain)...)
				assert.Nil(ba.Set(key, []byte(ref), pebble.NoSync))
			}
		}
	}
}
