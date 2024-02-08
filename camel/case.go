package camel

import (
	"strings"
	"unicode"
)

func Case(name string) string {
	first := true
	return strings.Map(func(r rune) rune {
		if first {
			first = false
			return unicode.ToLower(r)
		}
		return r
	}, name)
}
