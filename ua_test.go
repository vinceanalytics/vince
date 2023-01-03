package vince

import (
	"reflect"
	"testing"
)

func TestParseBot(t *testing.T) {
	ua := `Googlebot/2.1 (http://www.googlebot.com/bot.html)`
	expect := &botResult{
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

func TestParseOsUA(t *testing.T) {
	ua := "Mozilla/5.0 (AmigaOS; U; AmigaOS 1.3; en-US; rv:1.8.1.21) Gecko/20090303 SeaMonkey/1.1.15"
	got := parseOsUA(ua)
	want := &osResult{
		name:    "AmigaOS",
		version: "1.3",
	}
	if !reflect.DeepEqual(want, got) {
		t.Error("failed expectations")
	}
}

func Benchmark_parseOsUA(b *testing.B) {
	b.ReportAllocs()
	ua := "Mozilla/5.0 (AmigaOS; U; AmigaOS 1.3; en-US; rv:1.8.1.21) Gecko/20090303 SeaMonkey/1.1.15"
	for i := 0; i < b.N; i++ {
		_ = parseOsUA(ua)
	}
}

func TestParseDevice_camera(t *testing.T) {
	ua := `Mozilla/5.0 (Linux; U; Android 4.0; de-DE; EK-GC100 Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30`
	got := parseDeviceUA(ua)
	want := &deviceResult{
		model:   "Galaxy Camera 100",
		device:  "camera",
		company: "Samsung",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %#v got %#v", want, got)
	}
}

func Benchmark_parseDeviceUA(b *testing.B) {
	b.ReportAllocs()
	ua := `Mozilla/5.0 (Linux; U; Android 4.0; de-DE; EK-GC100 Build/IMM76D) AppleWebKit/534.30 (KHTML, like Gecko) Version/4.0 Mobile Safari/534.30`
	for i := 0; i < b.N; i++ {
		_ = parseDeviceUA(ua)
	}
}

func TestParseCLientUA(t *testing.T) {
	ua := "FeedDemon/4.5 (http://www.feeddemon.com/; Microsoft Windows)"
	got := parseClientUA(ua)
	want := &clientResult{kind: "Feed Reader", name: "FeedDemon", version: "4.5"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expect %#v got %#v", want, got)
	}
}

func Benchmark_parseClientUA(b *testing.B) {
	b.ReportAllocs()
	ua := "FeedDemon/4.5 (http://www.feeddemon.com/; Microsoft Windows)"
	for i := 0; i < b.N; i++ {
		_ = parseClientUA(ua)
	}
}
