package location

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	require.Equal(t, Country{Code: "IT", Name: "Italy", Flag: "🇮🇹"}, GetCountry("IT"))
	require.Equal(t, City{Name: "Rome", Flag: "🇮🇹"}, GetCity(3_169_070))
	require.Equal(t, Region{Name: "Lazio", Flag: "🇮🇹"}, GetRegion("IT-62"))
}
