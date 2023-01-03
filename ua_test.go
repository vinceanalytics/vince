package vince

import (
	"reflect"
	"testing"
)

func TestParseBot(t *testing.T) {
	ua := `Googlebot/2.1 (http://www.googlebot.com/bot.html)`
	expect := &botMatch{
		name:         "Googlebot",
		category:     "Search bot",
		url:          "http://www.google.com/bot.html",
		producerName: "Google Inc.",
		producerURL:  "http://www.google.com",
	}
	got := parseBotUA(ua)
	if !reflect.DeepEqual(got, expect) {
		t.Error("failed expectation")
	}
}

func Benchmark_parseBot(b *testing.B) {
	b.ReportAllocs()
	ua := `Googlebot/2.1 (http://www.googlebot.com/bot.html)`
	for i := 0; i < b.N; i++ {
		_ = parseBotUA(ua)
	}
}

func TestVendor(t *testing.T) {
	ua := "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0; Touch; MAARJS)"
	got := parseVendorUA(ua)
	if got != "Acer" {
		t.Errorf("expected Acre got %q", got)
	}
}

func Benchmark_parseVendor(b *testing.B) {
	b.ReportAllocs()
	ua := "Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.2; WOW64; Trident/6.0; Touch; MAARJS)"
	for i := 0; i < b.N; i++ {
		_ = parseVendorUA(ua)
	}
}
