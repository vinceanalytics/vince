package vince

import (
	"regexp"

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

func MatchRe(s string) *ReMatch {
	return &ReMatch{re: MustCompile(s)}
}

func MatchRe2(s string) *ReMatch {
	return &ReMatch{re2: MustCompile2(s)}
}
