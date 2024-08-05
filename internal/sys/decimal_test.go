package sys

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {
	a := fromFloat(6.897)
	b := toFloat(a)
	require.Equal(t, int64(6897), a)
	require.Equal(t, float64(6.897), b)
}
