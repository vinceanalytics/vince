package firewall

import (
	"net/http"

	"github.com/gernest/vince/remoteip"
)

type Wall interface {
	Allow(*http.Request) bool
}

type MatchFunc func(*http.Request) bool

func (f MatchFunc) Allow(r *http.Request) bool {
	return f(r)
}

type IP []string

func (f IP) Allow(r *http.Request) bool {
	if a := remoteip.Get(r); a != "" {
		for _, v := range f {
			if v == a {
				return true
			}
		}
	}
	return false
}

type List []Wall

func (f List) Allow(r *http.Request) bool {
	for _, v := range f {
		if !v.Allow(r) {
			return false
		}
	}
	return true
}

type Pass struct{}

func (f Pass) Allow(r *http.Request) bool {
	return true
}

func Negate(m Wall) Wall {
	return MatchFunc(func(r *http.Request) bool {
		return !m.Allow(r)
	})
}
