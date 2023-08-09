package be

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/vinceanalytics/vince/internal/must"
)

func TestParse(t *testing.T) {
	p := NewParser()
	r := must.Must(p.Parse("select * from `vince.io`"))()
	got := must.Must(
		json.MarshalIndent(
			must.Must(r.Plan.ToProto())(),
			"", "  ",
		),
	)()
	// os.WriteFile("testdata/simple_select.json", got, 0600)
	want := must.Must(os.ReadFile("testdata/simple_select.json"))()
	if !bytes.Equal(got, want) {
		t.Error("plan mismatch")
	}
}
