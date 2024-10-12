package ua2

import (
	"regexp"
	"strings"
	"sync"

	"github.com/VictoriaMetrics/fastcache"
	re2 "github.com/dlclark/regexp2"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/vinceanalytics/vince/fb"
	"github.com/vinceanalytics/vince/fb/ua"
	"github.com/vinceanalytics/vince/internal/models"
)

//go:generate go run gen/main.go device-detector/regexes/

var (
	cache = fastcache.New(32 << 20)
)

var agentPool = &sync.Pool{New: func() any {
	return new(ua.Agent)
}}

func Parse(s string, m *models.Model) {
	if buf := cache.Get(nil, []byte(s)); len(buf) > 0 {
		a := agentPool.Get().(*ua.Agent)
		a.Init(buf, flatbuffers.GetUOffsetT(buf))
		m.Device = a.DeviceBytes()
		m.Os = a.OsBytes()
		m.OsVersion = a.OsVersionBytes()
		m.Browser = a.BrowserBytes()
		m.BrowserVersion = a.BrowserVersionBytes()
		*a = ua.Agent{}
		agentPool.Put(a)
		return
	}
	parseUA(s, m)
	a := fb.SerializeAgent(m.Device, m.Os, m.OsVersion, m.Browser, m.BrowserVersion)
	cache.Set([]byte(s), a)
}

func parseUA(s string, m *models.Model) bool {
	if !containsLetter(s) {
		return false
	}
	if parseBotUA(s) {
		return true
	}
	parseDeviceUA(s, m)
	parseOsUA(s, m)
	parseClientUA(s, m)
	return false
}

func parseBotUA(ua string) (ok bool) {
	ok, _ = allBotsReStandardMatch().MatchString(ua)
	return
}

func parseOsUA(s string, m *models.Model) {
	if osAllRe.MatchString(s) {
		for _, e := range osAll {
			if e.re.MatchString(s) {
				var version string
				if e.version != "" {
					version = e.version
					if strings.HasPrefix(e.version, "$") {
						version = e.re.FirstSubmatch(s)
					}
				}
				m.Os = []byte(e.name)
				m.OsVersion = []byte(version)
				return
			}
		}
	}
}

var (
	devices = [][]byte{
		[]byte("Mobile"),
		[]byte("Tablet"),
		[]byte("Desktop"),
	}
)

func parseDeviceUA(s string, m *models.Model) {
	{
		// find cameras
		if deviceCameraAllRe.MatchString(s) {
			m.Device = devices[0]
			return
		}
	}
	{
		// find car browsers
		if deviceCarAllRe.MatchString(s) {
			m.Device = devices[1]
			return
		}
	}
	{
		// find consoles
		if deviceConsoleAllRe.MatchString(s) {
			m.Device = devices[2]
			return
		}
	}
	{
		// find mobiles
		if deviceMobileAllRe.MatchString(s) {
			m.Device = devices[0]
			return
		}
	}
	{
		// find notebooks
		if deviceNotebookAllRe.MatchString(s) {
			m.Device = devices[1]
			return
		}
	}
	{
		// find portable media player
		if devicePortableMediaPlayerAllRe.MatchString(s) {
			m.Device = devices[0]
			return
		}
	}
	{
		// find shell tv
		if deviceIsShellTV().MatchString(s) {
			m.Device = devices[2]
			return
		}
	}
	{
		// find tv
		if deviceIsTVRe().MatchString(s) {
			m.Device = devices[2]
			return
		}
	}
}

func parseClientUA(s string, m *models.Model) {
	{
		// find browsers
		if clientBrowserAllRe.MatchString(s) {
			parseClientBase(s, m, clientBrowserAll)
			return
		}
	}
	{
		// find feed readers
		if clientFeedReaderAllRe.MatchString(s) {
			parseClientBase(s, m, clientFeedReaderAll)
			return
		}
	}
	{
		// find libraries
		if clientLibraryAllRe.MatchString(s) {
			parseClientBase(s, m, clientLibraryAll)
			return
		}
	}
	{
		// find media players
		if clientMediaPlayerAllRe.MatchString(s) {
			parseClientBase(s, m, clientMediaPlayerAll)
			return
		}
	}
	{
		// find mobile apps
		if clientMobileAppAllRe.MatchString(s) {
			parseClientBase(s, m, clientMobileAppAll)
			return
		}
	}
	{
		if clientPimAllRe.MatchString(s) {
			parseClientBase(s, m, clientPimAll)
			return
		}
	}
}

func parseClientBase(s string, m *models.Model, ls []*clientRe) {
	for _, e := range ls {
		if e.re.MatchString(s) {
			m.Browser = []byte(e.name)
			if e.version == "$1" {
				m.BrowserVersion = []byte(e.re.FirstSubmatch(s))
			} else {
				m.Browser = []byte(e.name)
			}
			return
		}
	}
}

type clientRe struct {
	re      *ReMatch
	name    string
	version string
}

type osRe struct {
	re      *ReMatch
	name    string
	version string
}

func containsLetter(ua string) bool {
	for _, c := range ua {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}

type Re2Func = func() *re2.Regexp
type ReFunc = func() *regexp.Regexp

func MustCompile(s string) ReFunc {
	var r *regexp.Regexp
	return func() *regexp.Regexp {
		if r != nil {
			return r
		}
		r = regexp.MustCompile(s)
		return r
	}
}

func MustCompile2(s string) Re2Func {
	var r *re2.Regexp
	return func() *re2.Regexp {
		if r != nil {
			return r
		}
		r = re2.MustCompile(s, re2.IgnoreCase)
		return r
	}
}

type ReMatch struct {
	re  ReFunc
	re2 Re2Func
}

func (r *ReMatch) MatchString(s string) bool {
	if r.re != nil {
		return r.re().MatchString(s)
	}
	ok, _ := r.re2().MatchString(s)
	return ok
}

func (r *ReMatch) FirstSubmatch(s string) string {
	if r.re != nil {
		sub := r.re().FindStringSubmatch(s)
		if len(sub) > 1 {
			return sub[1]
		}
	}
	return ""
}

func MatchRe(s string) *ReMatch {
	return &ReMatch{re: MustCompile(s)}
}

func MatchRe2(s string) *ReMatch {
	return &ReMatch{re2: MustCompile2(s)}
}
