package timeseries

import (
	"github.com/segmentio/parquet-go"
)

// FIELD_TYPE represents a filterable column in parquet file.
type FIELD_TYPE uint

const (
	UNKNOWN FIELD_TYPE = iota
)

func (f FIELD_TYPE) String() string {
	return ""
}

func TypeFromString(str string) FIELD_TYPE {
	return UNKNOWN
}

type Field interface {
	Type() FIELD_TYPE
	Value() parquet.Value
}

type OP uint

const (
	BLOOM_EQ OP = iota
	BLOOM_NE
	BLOOM_AND_DICT_EQ
	DICT
)

type FILTER struct {
	Field Field
	Op    OP
}

func (f *FILTER) Match(a any) bool {
	return true
}
