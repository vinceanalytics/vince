package mutex

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/kase"
)

func TestMutex_extract(t *testing.T) {
	suite.Run(t, &kase.Kase[uint64]{
		Add:     Add,
		Extract: Extract,
		Source:  []uint64{1, 2, 3, 4, 5},
	})
}
