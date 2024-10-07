package ua2

import (
	"regexp"
	"strings"

	re2 "github.com/dlclark/regexp2"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

func Parse(s string) (a *v1.Agent) {
	if ua := parseUA(s); ua != nil {
		a = &v1.Agent{}
		a.Bot = ua.bot != nil
		if ua.os != nil {
			a.Os = ua.os.name
			a.OsVersion = ua.os.version
		}
		if ua.client != nil {
			a.Browser = ua.client.name
			a.BrowserVersion = ua.client.version
		}
		if ua.device != nil {
			a.Device = ua.device.device
		}
	}
	return
}

func parseUA(s string) *deviceInfo {
	if !containsLetter(s) {
		return nil
	}
	return &deviceInfo{
		ua:     s,
		device: parseDeviceUA(s),
		client: parseClientUA(s),
		os:     parseOsUA(s),
		bot:    parseBotUA(s),
	}
}

func parseBotUA(ua string) *botResult {
	if ok, _ := allBotsReStandardMatch().MatchString(ua); ok {
		for _, m := range botsReList {
			if m.re.MatchString(ua) {
				return &botResult{
					name:         m.name,
					category:     m.category,
					url:          m.url,
					producerName: m.producerName,
					producerURL:  m.producerURL,
				}
			}
		}
		return nil
	}
	return nil
}

func parseVendorUA(s string) string {
	if vendorAllRe.MatchString(s) {
		for _, r := range vendorAll {
			if r.re.MatchString(s) {
				return r.name
			}
		}
	}
	return ""
}

func parseOsUA(s string) *osResult {
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
				return &osResult{
					name:    e.name,
					version: version,
				}
			}
		}
	}
	return nil
}

func parseDeviceUA(s string) *deviceResult {
	{
		// find cameras
		if deviceCameraAllRe.MatchString(s) {
			return parseDeviceBase("Mobile", s, deviceCameraAll)
		}
	}
	{
		// find car browsers
		if deviceCarAllRe.MatchString(s) {
			return parseDeviceBase("Tablet", s, deviceCarAll)
		}
	}
	{
		// find consoles
		if deviceConsoleAllRe.MatchString(s) {
			return parseDeviceBase("Desktop", s, deviceConsoleAll)
		}
	}
	{
		// find mobiles
		if deviceMobileAllRe.MatchString(s) {
			return parseDeviceBase("Mobile", s, deviceMobileAll)
		}
	}
	{
		// find notebooks
		if deviceNotebookAllRe.MatchString(s) {
			return parseDeviceBase("Tablet", s, deviceNotebookAll)
		}
	}
	{
		// find portable media player
		if devicePortableMediaPlayerAllRe.MatchString(s) {
			return parseDeviceBase("Mobile", s, devicePortableMediaPlayerAll)
		}
	}
	{
		// find shell tv
		if deviceIsShellTV().MatchString(s) {
			return parseDeviceBase("Desktop", s, deviceShellAll)
		}
	}
	{
		// find tv
		if deviceIsTVRe().MatchString(s) {
			return parseDeviceBase("Desktop", s, deviceTVAll)
		}
	}
	return nil
}

func parseDeviceBase(kind, s string, ls []*deviceRe) *deviceResult {
	for _, e := range ls {
		if e.re.MatchString(s) {
			d := &deviceResult{
				model:   e.model,
				company: e.company,
				device:  kind,
			}
			if len(e.models) > 0 {
				for _, m := range e.models {
					if m.re.MatchString(s) {
						d.model = m.model
						d.model = strings.Replace(d.model, "$1", m.re.FirstSubmatch(s), -1)
					}
				}
			}
			return d
		}
	}
	return nil
}

func parseClientUA(s string) *clientResult {
	{
		// find browsers
		if clientBrowserAllRe.MatchString(s) {
			return parseClientBase(s, clientBrowserAll)
		}
	}
	{
		// find feed readers
		if clientFeedReaderAllRe.MatchString(s) {
			return parseClientBase(s, clientFeedReaderAll)
		}
	}
	{
		// find libraries
		if clientLibraryAllRe.MatchString(s) {
			return parseClientBase(s, clientLibraryAll)
		}
	}
	{
		// find media players
		if clientMediaPlayerAllRe.MatchString(s) {
			return parseClientBase(s, clientMediaPlayerAll)
		}
	}
	{
		// find mobile apps
		if clientMobileAppAllRe.MatchString(s) {
			return parseClientBase(s, clientMobileAppAll)
		}
	}
	{
		if clientPimAllRe.MatchString(s) {
			return parseClientBase(s, clientPimAll)
		}
	}
	return nil
}

func parseClientBase(s string, ls []*clientRe) *clientResult {
	for _, e := range ls {
		if e.re.MatchString(s) {
			d := &clientResult{
				kind:    e.kind,
				name:    e.name,
				version: e.version,
			}
			if strings.Contains(d.version, "$1") {
				d.version = strings.Replace(d.version, "$1", e.re.FirstSubmatch(s), -1)
			}
			return d
		}
	}
	return nil
}

type botRe struct {
	re           *ReMatch
	name         string
	category     string
	url          string
	producerName string
	producerURL  string
}

type botResult struct {
	name         string
	category     string
	url          string
	producerName string
	producerURL  string
}

type clientRe struct {
	re      *ReMatch
	name    string
	version string
	kind    string
	url     string
	engine  *clientEngine
}

type clientEngine struct {
	def      string
	versions map[string]string
}

type clientResult struct {
	kind    string
	name    string
	version string
}

type deviceRe struct {
	re      *ReMatch
	model   string
	device  string
	company string
	models  []*deviceModel
}

type deviceModel struct {
	re    *ReMatch
	model string
}

type deviceResult struct {
	model   string
	device  string
	company string
}

type osRe struct {
	re      *ReMatch
	name    string
	version string
}
type osResult struct {
	name    string
	version string
}

type vendorRe struct {
	re   *ReMatch
	name string
}

type vendorResult struct {
	name string
}

type deviceInfo struct {
	ua     string
	device *deviceResult
	client *clientResult
	os     *osResult
	bot    *botResult
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
