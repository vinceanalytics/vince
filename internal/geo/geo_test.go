package geo

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/location"
	"github.com/vinceanalytics/vince/internal/models"
)

func TestGet(t *testing.T) {
	m := new(models.Model)
	require.NoError(t, UpdateCity(net.ParseIP("1.0.16.0"), m))
	require.Equal(t, location.City{Name: "Chiyoda", Flag: "🇯🇵"}, location.GetCity(m.City))
	require.Equal(t, location.Country{Alpha: "JPN", Code: "JP", Name: "Japan", Flag: "🇯🇵"}, location.GetCountry(string(m.Country)))
	require.Equal(t, location.Region{Name: "Tokyo", Flag: "🇯🇵"}, location.GetRegion(m.Subdivision1Code))
}

func BenchmarkGet(b *testing.B) {
	ip := net.ParseIP("1.0.16.0")
	m := new(models.Model)

	for range b.N {
		UpdateCity(ip, m)
	}
}
