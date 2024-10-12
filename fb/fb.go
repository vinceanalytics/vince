package fb

import (
	"bytes"
	"slices"
	"sort"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/vinceanalytics/vince/fb/geo"
	"github.com/vinceanalytics/vince/fb/ua"
	"github.com/vinceanalytics/vince/internal/roaring"
)

//go:generate flatc --go geo.fbs ua.fbs

func SerializeAgent(device, os, osVersion, browser, browserVersion []byte) []byte {
	size := len(device) + len(os) + len(osVersion) + len(browser) + len(browserVersion)
	if size == 0 {
		b := flatbuffers.NewBuilder(0)
		ua.AgentStart(b)
		b.Finish(ua.AgentEnd(b))
		return b.FinishedBytes()
	}
	b := flatbuffers.NewBuilder(size)
	deviceData := buildBytes(b, device, ua.AgentStartDeviceVector)
	osData := buildBytes(b, os, ua.AgentStartOsVector)
	osVerData := buildBytes(b, osVersion, ua.AgentStartOsVersionVector)
	broData := buildBytes(b, browser, ua.AgentStartBrowserVector)
	broVerData := buildBytes(b, browserVersion, ua.AgentStartBrowserVersionVector)
	ua.AgentStart(b)
	ua.AgentAddDevice(b, deviceData)
	ua.AgentAddOs(b, osData)
	ua.AgentAddOsVersion(b, osVerData)
	ua.AgentAddBrowser(b, broData)
	ua.AgentAddBrowserVersion(b, broVerData)
	b.Finish(ua.AgentEnd(b))
	return b.FinishedBytes()
}

func Serialize(geonames []string, country [4][]string, region [2][]string, city, cityCode []byte) []byte {
	b := flatbuffers.NewBuilder(30 << 20)

	off, names := buildStrings(b, nil, geonames, geo.GeoStartNamesVector)
	off, countryAlpha := buildStrings(b, off, country[0], geo.CountryStartAlphaVector)
	off, countryCode := buildStrings(b, off, country[1], geo.CountryStartCodeVector)
	off, countryFlag := buildStrings(b, off, country[2], geo.CountryStartFlagVector)
	off, countryName := buildStrings(b, off, country[3], geo.CountryStartNameVector)

	off, regionCode := buildStrings(b, off, region[0], geo.RegionStartCodeVector)
	_, regionName := buildStrings(b, off, region[1], geo.RegionStartNameVector)

	cityData := buildBytes(b, city, geo.LocationStartCityVector)
	cityCodeData := buildBytes(b, cityCode, geo.LocationStartCityCodeVector)

	trData := buildBytes(b, newTrBSI(region[1]), geo.LocationStartTranslateVector)

	geo.GeoStart(b)
	geo.GeoAddNames(b, names)
	geoFull := geo.GeoEnd(b)

	geo.CountryStart(b)
	geo.CountryAddAlpha(b, countryAlpha)
	geo.CountryAddCode(b, countryCode)
	geo.CountryAddFlag(b, countryFlag)
	geo.CountryAddName(b, countryName)
	countryFull := geo.CountryEnd(b)

	geo.RegionStart(b)
	geo.RegionAddCode(b, regionCode)
	geo.RegionAddName(b, regionName)
	regionFull := geo.RegionEnd(b)

	geo.LocationStart(b)
	geo.LocationAddGeo(b, geoFull)
	geo.LocationAddCountry(b, countryFull)
	geo.LocationAddRegion(b, regionFull)
	geo.LocationAddCity(b, cityData)
	geo.LocationAddCityCode(b, cityCodeData)
	geo.LocationAddTranslate(b, trData)

	location := geo.LocationEnd(b)
	b.Finish(location)
	return b.FinishedBytes()
}

func newTrBSI(names []string) []byte {
	b := roaring.NewDefaultBSI()
	tr := newTr(names)
	for i := range tr {
		b.SetValue(uint64(i), int64(tr[i]))
	}
	return b.ToBuffer()
}

type translate struct {
	str []string
	idx []int
}

func newTr(str []string) []int {
	o := make([]int, len(str))
	for i := range o {
		o[i] = i
	}
	sort.Sort(&translate{str: str, idx: o})
	return o
}

func (t *translate) Len() int {
	return len(t.str)
}
func (t *translate) Less(i, j int) bool {
	return t.str[t.idx[i]] < t.str[t.idx[j]]
}

func (t *translate) Swap(i, j int) {
	t.idx[i], t.idx[j] = t.idx[j], t.idx[i]
}

func buildStrings(b *flatbuffers.Builder, offsets []flatbuffers.UOffsetT, names []string, start func(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT) ([]flatbuffers.UOffsetT, flatbuffers.UOffsetT) {
	offsets = slices.Grow(offsets[:0], len(names))
	for i := range names {
		offsets = append(offsets, b.CreateString(names[i]))
	}
	start(b, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offsets[i])
	}
	return offsets, b.EndVector(len(names))
}

func buildBytes(b *flatbuffers.Builder, names []byte, start func(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT) flatbuffers.UOffsetT {
	start(b, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		b.PrependByte(names[i])
	}
	return b.EndVector(len(names))
}

type Location struct {
	root                                                   *geo.Location
	Geo                                                    *Str
	CountryAddAlpha, CountryCode, CountryFlag, CountryName *Str
	RegionName, RegionCode                                 *Str
	City, CityCode, translate                              *roaring.BSI
	tr                                                     *Str
}

type Str struct {
	get  func(j int) []byte
	size int
}

func New(data []byte) *Location {
	root := geo.GetRootAsLocation(data, 0)
	var g Location
	g.root = root
	geo := root.Geo(nil)
	country := root.Country(nil)
	region := root.Region(nil)
	g.Geo = &Str{get: geo.Names, size: geo.NamesLength()}

	g.CountryAddAlpha = &Str{get: country.Alpha, size: country.AlphaLength()}
	g.CountryCode = &Str{get: country.Code, size: country.CodeLength()}
	g.CountryFlag = &Str{get: country.Flag, size: country.FlagLength()}
	g.CountryName = &Str{get: country.Name, size: country.NameLength()}

	g.RegionName = &Str{get: region.Name, size: region.NameLength()}
	g.RegionCode = &Str{get: region.Code, size: region.CodeLength()}

	g.City = roaring.NewBSIFromBuffer(root.CityBytes())
	g.CityCode = roaring.NewBSIFromBuffer(root.CityCodeBytes())
	g.translate = roaring.NewBSIFromBuffer(root.TranslateBytes())
	g.tr = &Str{
		size: g.RegionCode.size,
		get: func(j int) []byte {
			v, ok := g.translate.GetValue(uint64(j))
			if !ok {
				return []byte{}
			}
			return g.RegionName.get(int(v))
		},
	}
	return &g
}

func (g *Location) Translate(region string) []byte {
	i, ok := g.tr.Search([]byte(region))
	if !ok {
		return nil
	}
	idx, _ := g.translate.GetValue(uint64(i))
	return g.RegionCode.Get(int(idx))
}

func (g *Str) Get(i int) []byte {
	return g.get(i)
}

func (g *Str) Search(key []byte) (int, bool) {
	n := g.size
	// Define cmp(x[-1], target) < 0 and cmp(x[n], target) >= 0 .
	// Invariant: cmp(x[i - 1], target) < 0, cmp(x[j], target) >= 0.
	i, j := 0, n
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j

		if bytes.Compare(g.get(h), key) < 0 {
			i = h + 1 // preserves cmp(x[i - 1], target) < 0
		} else {
			j = h // preserves cmp(x[j], target) >= 0
		}
	}
	// i == j, cmp(x[i-1], target) < 0, and cmp(x[j], target) (= cmp(x[i], target)) >= 0  =>  answer is i.
	return i, i < n && bytes.Equal(g.get(i), key)
}
