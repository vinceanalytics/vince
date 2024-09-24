package sets

import (
	"testing"

	"github.com/gernest/rbf/dsl/kase"
	"github.com/stretchr/testify/suite"
)

func TestSets_extract(t *testing.T) {
	suite.Run(t, &kase.Kase[[]uint64]{
		Add:     Add,
		Extract: Extract,
		Source:  [][]uint64{{1}, {2, 6}, {3}, {0, 2, 4}, {5}},
	})
}
