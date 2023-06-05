package referrer

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeFavicon(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ok := ServeFavicon("Google", w, r)
	if !ok {
		t.Fatal("expected to succeed")
	}
}
