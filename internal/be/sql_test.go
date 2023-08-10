package be

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/vinceanalytics/vince/internal/must"
)

func TestParse_simple_select(t *testing.T) {
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

func TestParse_simple_select_columns(t *testing.T) {
	p := NewParser()
	r := must.Must(p.Parse("select name,bounce from `vince.io`"))()
	got := must.Must(
		json.MarshalIndent(
			must.Must(r.Plan.ToProto())(),
			"", "  ",
		),
	)()
	// os.WriteFile("testdata/simple_select_columns.json", got, 0600)
	want := must.Must(os.ReadFile("testdata/simple_select_columns.json"))()
	if !bytes.Equal(got, want) {
		t.Error("plan mismatch")
	}
}

func TestParse_simple_where(t *testing.T) {
	p := NewParser()
	r := must.Must(p.Parse("select * from `vince.io` where name='pageview'"))()
	got := must.Must(
		json.MarshalIndent(
			must.Must(r.Plan.ToProto())(),
			"", "  ",
		),
	)()
	os.WriteFile("testdata/simple_select_where.json", got, 0600)
	want := must.Must(os.ReadFile("testdata/simple_select_where.json"))()
	if !bytes.Equal(got, want) {
		t.Error("plan mismatch")
	}
	t.Error()

}
