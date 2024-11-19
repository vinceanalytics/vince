package location

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	db := New()
	require.Equal(t, Country{Alpha: "ITA", Code: "IT", Name: "Italy", Flag: "🇮🇹"}, db.GetCountry("IT"))
	require.Equal(t, City{Name: "Rome", Flag: "🇮🇹"}, db.GetCity(3_169_070))
	require.Equal(t, uint32(3_169_070), db.GetCityCode("IT", "Rome"))
	require.Equal(t, Region{Name: "Lazio", Flag: "🇮🇹"}, db.GetRegion([]byte("IT-62")))
	require.Equal(t, []byte("IT-62"), db.GetRegionCode("Lazio"))
}

func BenchmarkGetCityCode(b *testing.B) {
	db := New()
	defer db.Close()

	for range b.N {
		db.GetCityCode("IT", "Rome")
	}
}
func BenchmarkGetRegionCode(b *testing.B) {
	db := New()
	defer db.Close()

	for range b.N {
		db.GetRegionCode("Lazio")
	}
}
