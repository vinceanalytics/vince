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
	lo := location.New()
	defer lo.Close()

	g, err := New(lo, t.TempDir())
	require.Nil(t, err)
	require.NoError(t, g.UpdateCity(net.ParseIP("1.0.16.0"), m))
	require.Equal(t, location.City{Name: "Chiyoda", Flag: "🇯🇵"}, lo.GetCity(m.City))
	require.Equal(t, location.Country{Alpha: "JPN", Code: "JP", Name: "Japan", Flag: "🇯🇵"}, lo.GetCountry(string(m.Country)))
	require.Equal(t, location.Region{Name: "Tokyo", Flag: "🇯🇵"}, lo.GetRegion(m.Subdivision1Code))
}

func BenchmarkGet(b *testing.B) {

	lo := location.New()
	defer lo.Close()
	g, err := New(lo, b.TempDir())
	require.Nil(b, err)
	b.ResetTimer()

	ip := net.ParseIP("1.0.16.0")
	m := new(models.Model)

	for range b.N {
		g.UpdateCity(ip, m)
	}
}
