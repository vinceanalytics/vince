package boolean

import (
	"testing"

	"github.com/gernest/rbf/dsl/kase"
	"github.com/stretchr/testify/suite"
)

func TestBoolean_extract(t *testing.T) {
	suite.Run(t, &kase.Kase[bool]{
		Add:     Add,
		Extract: Extract,
		Source:  []bool{true, false, true, false, true},
	})
}
