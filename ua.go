package vince

import (
	"regexp"
	"strings"

	re2 "github.com/dlclark/regexp2"
)

func parseBotUA(ua string) *botMatch {
	if ok, _ := allBotsReStandardMatch().MatchString(ua); ok {
		for _, m := range botsReList {
			if m.re.MatchString(ua) {
				return &botMatch{
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
					if strings.HasPrefix(e.version, "$") {
						sub := e.re.re().FindStringSubmatch(s)
						if len(sub) > 1 {
							version = sub[1]
						}
					} else {
						version = e.version
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
			return parseDeviceBase(s, deviceCameraAll)
		}
	}
	{
		// find car browsers
		if deviceCarAllRe.MatchString(s) {
			return parseDeviceBase(s, deviceCarAll)
		}
	}
	{
		// find consoles
		if deviceConsoleAllRe.MatchString(s) {
			return parseDeviceBase(s, deviceConsoleAll)
		}
	}
	{
		// find mobiles
		if deviceMobileAllRe.MatchString(s) {
			return parseDeviceBase(s, deviceMobileAll)
		}
	}
	{
		// find notebooks
		if deviceNotebookAllRe.MatchString(s) {
			return parseDeviceBase(s, deviceNotebookAll)
		}
	}
	{
		// find portable media player
		if devicePortableMediaPlayerAllRe.MatchString(s) {
			return parseDeviceBase(s, devicePortableMediaPlayerAll)
		}
	}
	{
		// find shell tv
		if deviceIsShellTV().MatchString(s) {
			return parseDeviceBase(s, deviceShellAll)
		}
	}
	{
		// find tv
		if deviceIsTVRe().MatchString(s) {
			return parseDeviceBase(s, deviceTVAll)
		}
	}
	return nil
}

func parseDeviceBase(s string, ls []*deviceRe) *deviceResult {
	for _, e := range ls {
		if e.re.MatchString(s) {
			d := &deviceResult{
				model:   e.model,
				company: e.company,
				device:  e.device,
			}
			if len(e.models) > 0 {
				for _, m := range e.models {
					if m.re.MatchString(s) {
						d.model = m.model
						if strings.Contains(d.model, "$1") {
							d.model = strings.Replace(d.model, "$1", m.re.FirstSubmatch(s), -1)
						}
					}
				}
			}
			return d
		}
	}
	return nil
}
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

type Re2Func = func() *re2.Regexp
type ReFunc = func() *regexp.Regexp

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
