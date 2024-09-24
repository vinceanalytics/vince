package bsi

import (
	"math"
	"testing"

	"github.com/gernest/rbf/dsl/kase"
	"github.com/gernest/roaring/shardwidth"
	"github.com/stretchr/testify/suite"
)

const (
	// NormalNaN is a quiet NaN. This is also math.NaN().
	NormalNaN uint64 = 0x7ff8000000000001

	// StaleNaN is a signaling NaN, due to the MSB of the mantissa being 0.
	// This value is chosen with many leading 0s, so we have scope to store more
	// complicated values in the future. It is 2 rather than 1 to make
	// it easier to distinguish from the NormalNaN by a human when debugging.
	StaleNaN uint64 = 0x7ff0000000000002
)

func TestBSI_extract(t *testing.T) {
	suite.Run(t, &kase.Kase[int64]{
		Add:     Add,
		Extract: Extract,
		Source:  []int64{int64(NormalNaN), int64(StaleNaN), math.MaxInt64, -shardwidth.ShardWidth, 0},
	})
}
