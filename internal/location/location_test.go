package location

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	require.Equal(t, Country{Code: "IT", Name: "Italy", Flag: "🇮🇹"}, GetCountry("IT"))
	require.Equal(t, City{Name: "Rome", Flag: "🇮🇹"}, GetCity(3_169_070))
	require.Equal(t, uint32(3_169_070), GetCityCode("IT", "Rome"))
	require.Equal(t, Region{Name: "Lazio", Flag: "🇮🇹"}, GetRegion("IT-62"))
	require.Equal(t, "IT-62", GetRegionCode("Lazio"))
}

func BenchmarkGetCityCode(b *testing.B) {
	for range b.N {
		GetCityCode("IT", "Rome")
	}
}
func BenchmarkGetRegionCode(b *testing.B) {
	for range b.N {
		GetRegionCode("Lazio")
	}
}
