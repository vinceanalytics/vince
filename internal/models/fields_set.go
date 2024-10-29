package models

import (
	"math/bits"
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
	for i := Field_domain; i <= Field_duration; i++ {
		if v.Test(i) {
			o = append(o, i)
		}
	}
	return
}
