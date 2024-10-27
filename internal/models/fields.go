package models

import (
	"github.com/bits-and-blooms/bitset"
)

const (
	MutexFieldSize       = Field_session + 1
	TranslatedFieldsSize = Field_subdivision2_code + 1
	SearchFieldSize      = Field_city + 1
	BSIFieldsSize        = Field_duration - Field_timestamp + 1
	AllFields            = Field_duration + 1
)

func (f Field) Mutex() byte {
	return byte(f)
}

func (f Field) BSI() byte {
	return byte(f - Field_timestamp)
}

func BSI(i int) Field {
	return Field(i) + Field_timestamp
}

func Mutex(i int) Field {
	return Field(i)
}

// Data returns a set of all data fields.
func Data() *bitset.BitSet {
	s := new(bitset.BitSet)
	s.Set(uint(Field_bounce))
	s.Set(uint(Field_duration))
	s.Set(uint(Field_event))
	s.Set(uint(Field_id))
	s.Set(uint(Field_session))
	s.Set(uint(Field_view))
	return s
}

func DataForMetrics(m ...string) *bitset.BitSet {
	s := new(bitset.BitSet)
	for _, v := range m {
		switch v {
		case "visitors":
			s.Set(uint(Field_id))
		case "visits":
			s.Set(uint(Field_session))
		case "pageviews":
			s.Set(uint(Field_view))
		case "views_per_visit":
			s.Set(uint(Field_view))
			s.Set(uint(Field_session))
		case "bounce_rate":
			s.Set(uint(Field_bounce))
			s.Set(uint(Field_session))
		case "visit_duration":
			s.Set(uint(Field_duration))
			s.Set(uint(Field_session))
		case "events":
			s.Set(uint(Field_event))
		}
	}
	return s
}

func EachField(bs *bitset.BitSet, f func(f Field)) {
	_, all := bs.NextSetMany(0, make([]uint, AllFields))
	for _, v := range all {
		f(Field(v))
	}
}
