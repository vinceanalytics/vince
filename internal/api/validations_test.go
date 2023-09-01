package api

import (
	"strings"
	"testing"

	"github.com/bufbuild/protovalidate-go"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
)

func TestCreateSiteRequest(t *testing.T) {
	v, err := protovalidate.New()
	if err != nil {
		t.Fatal()
	}
	t.Run("domain is required", func(t *testing.T) {
		err := v.Validate(&v1.CreateSiteRequest{})
		if err == nil || !strings.Contains(err.Error(), "required") {
			t.Error("expected an error ")
		}
	})
	t.Run("reject invalid hostname", func(t *testing.T) {
		err := v.Validate(&v1.CreateSiteRequest{
			Domain: "https://vinceanalytics.github.com",
		})
		if err == nil || !strings.Contains(err.Error(), "string.hostname") {
			t.Error("expected an error ")
		}
	})
	t.Run("accept valid hostname", func(t *testing.T) {
		err := v.Validate(&v1.CreateSiteRequest{
			Domain: "vinceanalytics.github.com",
		})
		if err != nil {
			t.Error(err)
		}
	})
}
