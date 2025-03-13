package models

import (
	"math/bits"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
)

type BitSet uint64

func (v BitSet) Test(i Field) bool {
	return v&(1<<uint64(i&63)) != 0
}

func (v *BitSet) Set(i Field) {
	*v |= (1 << uint64(i))
}

func (v *BitSet) Unset(i int) {
	*v &= ^(1 << uint64(i))
}

func (v BitSet) Or(other BitSet) BitSet {
	return v | other
}

func (v BitSet) Len() int {
	return bits.OnesCount64(uint64(v))
}

func (v BitSet) All() (o []Field) {
	o = make([]Field, 0, v.Len())
	for i := v1.Field_domain; i <= v1.Field_duration; i++ {
		if v.Test(i) {
			o = append(o, i)
		}
	}
	return
}
