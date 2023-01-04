package geoip

import (
	"net"
	"testing"
)

func TestSample(t *testing.T) {
	ip := net.ParseIP("81.2.69.142")
	city, err := Get().City(ip)
	if err != nil {
		t.Fatal(err)
	}
	if city.Country.IsoCode != "GB" {
		t.Error("failed expectations")
	}
}
