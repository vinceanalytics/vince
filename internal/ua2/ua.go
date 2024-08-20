package ua2

import (
	"regexp"

	re2 "github.com/dlclark/regexp2"
)

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

type Agent struct {
	Os             string
	OsVersion      string
	Browser        string
	BrowserVersion string
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
