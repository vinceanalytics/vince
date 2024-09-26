package ro2

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/vinceanalytics/vince/internal/alicia"
	"github.com/vinceanalytics/vince/internal/roaring/roaring64"
)

// We know fields before hand
type Data [alicia.SUB2_CODE]*roaring64.BSI

var dataPool = &sync.Pool{
	New: func() any {
		var d Data
		return &d
	},
}

func NewData() *Data {
	return dataPool.Get().(*Data)
}

func (d *Data) Release() {
	clear(d[:])
	dataPool.Put(d)
}

func (d *Data) mustGet(i alicia.Field) *roaring64.BSI {
	i--
	if d[i] == nil {
		d[i] = roaring64.NewDefaultBSI()
	}
	return d[i]
}

func (d *Data) get(i alicia.Field) *roaring64.BSI {
	i--
	return d[i]
}

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
