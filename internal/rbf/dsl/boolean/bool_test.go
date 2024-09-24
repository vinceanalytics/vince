package boolean

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vinceanalytics/vince/internal/rbf/dsl/kase"
)

func TestBoolean_extract(t *testing.T) {
	suite.Run(t, &kase.Kase[bool]{
		Add:     Add,
		Extract: Extract,
		Source:  []bool{true, false, true, false, true},
	})
}
