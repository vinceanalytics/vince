package encoding

import (
	"bytes"
	"cmp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestCompare(t *testing.T) {
	a := Bitmap(0, 1730023200000, models.Field_domain)
	b := Bitmap(0, 1730026800000, models.Field_domain)
	require.Equal(t, -1, cmp.Compare(1730023200000, 1730026800000))
	require.Equal(t, -1, bytes.Compare(a, b))

	sa, va := Component(a)
	require.Equal(t, uint64(0), sa)
	require.Equal(t, uint64(1730023200000), va)

	sb, vb := Component(b)
	require.Equal(t, uint64(0), sb)
	require.Equal(t, uint64(1730026800000), vb)
}
