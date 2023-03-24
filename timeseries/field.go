package timeseries

import (
	"github.com/apache/arrow/go/v12/arrow/compute"
	"github.com/segmentio/parquet-go"
)

type OP uint

const (
	BLOOM_EQ OP = iota
	BLOOM_NE
	EQ
	NE
)

func (op OP) String() string {
	switch op {
	case EQ, BLOOM_EQ:
		return "equal"
	case NE, BLOOM_NE:
		return "not_equal"
	default:
		return ""
	}
}

type FILTER struct {
	Scalar  *compute.ScalarDatum
	Field   string
	Parquet parquet.Value
	Op      OP
}
