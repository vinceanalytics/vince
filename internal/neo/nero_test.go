package neo

import (
	"testing"
)

func TestT(t *testing.T) {
	for _, f := range FieldToArrowType {
		println(f.String())
	}

}
