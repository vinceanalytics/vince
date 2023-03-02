package timeseries

import (
	"testing"
)

func TestParseFilters(t *testing.T) {
	t.Run("simple positive", func(t *testing.T) {
		expect := "event:name is pageview"
		got := parseFilters("event:name==pageview").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("simple negative", func(t *testing.T) {
		expect := "event:name is_not pageview"
		got := parseFilters("event:name!=pageview").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("whitespace is trimmed", func(t *testing.T) {
		expect := "event:name is pageview"
		got := parseFilters(" event:name == pageview ").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("wildcard", func(t *testing.T) {
		expect := "event:page matches /blog/post-*"
		got := parseFilters("event:page==/blog/post-*").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("negative wildcard", func(t *testing.T) {
		expect := "event:page matches_not /blog/post-*"
		got := parseFilters("event:page!=/blog/post-*").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("custom event goal", func(t *testing.T) {
		expect := "event:goal[:is :event \"Signup\"]"
		got := parseFilters("event:goal==Signup").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("pageview goal", func(t *testing.T) {
		expect := "event:goal[:is :page \"/blog\"]"
		got := parseFilters("event:goal==Visit /blog").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("member", func(t *testing.T) {
		expect := "visit:country [ FR, GB, DE ]"
		got := parseFilters("visit:country==FR|GB|DE").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("member + wildcard", func(t *testing.T) {
		expect := "event:page matches /blog**|/newsletter|/*/"
		got := parseFilters("event:page==/blog**|/newsletter|/*/").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})
	t.Run("combined with \";\"", func(t *testing.T) {
		expect := "event:page matches /blog**|/newsletter|/*/\nvisit:country [ FR, GB, DE ]"
		got := parseFilters("event:page==/blog**|/newsletter|/*/ ; visit:country==FR|GB|DE").String()
		if expect != got {
			t.Errorf("expected %q got %q", expect, got)
		}
	})

	t.Run("escaping pipe character", func(t *testing.T) {
		expect := "utm_campaign is campaign | 1"
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
		expect := "event:page matches /**\\|page|/other/page"
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
