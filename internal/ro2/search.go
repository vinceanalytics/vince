package ro2

import (
	"strings"
	"unicode/utf8"
)

func cleanRe(re string) string {
	re = strings.TrimPrefix(re, "~")
	re = strings.TrimSuffix(re, "$")
	return re
}

func searchPrefix(source []byte) (prefix []byte, exact bool) {
	for i := range source {
		if special(source[i]) {
			return source[:i], false
		}
	}
	return source, true
}

// Bitmap used by func special to check whether a character needs to be escaped.
var specialBytes [16]byte

// special reports whether byte b needs to be escaped by QuoteMeta.
func special(b byte) bool {
	return b < utf8.RuneSelf && specialBytes[b%16]&(1<<(b/16)) != 0
}

func init() {
	for _, b := range []byte(`\.+*?()|[]{}^$`) {
		specialBytes[b%16] |= 1 << (b / 16)
	}
}
