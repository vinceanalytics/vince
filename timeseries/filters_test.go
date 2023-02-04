package timeseries

import "testing"

func TestFilters(t *testing.T) {
	t.Run("escaping pipe character", func(t *testing.T) {
		expect := "utm_campaign == campaign | 1"
		got := parseFilters("utm_campaign==campaign \\| 1").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("escaping pipe character in member filter", func(t *testing.T) {
		expect := "utm_campaign [ campaign | 1, campaign | 2 ]"
		got := parseFilters("utm_campaign==campaign \\| 1|campaign \\| 2").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("keeps escape characters in member + wildcard filter", func(t *testing.T) {
		expect := "event:page == /**\\|page|/other/page"
		got := parseFilters("event:page==/**\\|page|/other/page").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("gracefully fails to parse garbage", func(t *testing.T) {
		expect := ""
		got := parseFilters("bfg10309\uff1cs1\ufe65s2\u02bas3\u02b9hjl10309").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
}
