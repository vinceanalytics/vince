package geo

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vinceanalytics/vince/internal/location"
)

func TestGet(t *testing.T) {
	info, err := Get(net.ParseIP("1.0.16.0"))
	require.NoError(t, err)
	require.Equal(t, Info{CountryCode: "JP", SubDivision1Code: "JP-13", SubDivision2Code: "", CityGeonameID: 0xa6c578}, info)
	require.Equal(t, location.City{Name: "Chiyoda", Flag: "🇯🇵"}, location.GetCity(info.CityGeonameID))
	require.Equal(t, location.Country{Code: "JP", Name: "Japan", Flag: "🇯🇵"}, location.GetCountry(info.CountryCode))
	require.Equal(t, location.Region{Name: "Tokyo", Flag: "🇯🇵"}, location.GetRegion(info.SubDivision1Code))
}

func BenchmarkGet(b *testing.B) {
	ip := net.ParseIP("1.0.16.0")
	for range b.N {
		Get(ip)
	}
}
