package geo

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/location"
)

func TestGet(t *testing.T) {
	m := new(v1.Model)
	require.NoError(t, UpdateCity(net.ParseIP("1.0.16.0"), m))
	require.Equal(t, location.City{Name: "Chiyoda", Flag: "🇯🇵"}, location.GetCity(m.City))
	require.Equal(t, location.Country{Code: "JP", Name: "Japan", Flag: "🇯🇵"}, location.GetCountry(string(m.Country)))
	require.Equal(t, location.Region{Name: "Tokyo", Flag: "🇯🇵"}, location.GetRegion(m.Subdivision1Code))
}

func BenchmarkGet(b *testing.B) {
	ip := net.ParseIP("1.0.16.0")
	m := new(v1.Model)

	for range b.N {
		UpdateCity(ip, m)
	}
}
