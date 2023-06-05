package config

import (
	"testing"
)

func TestCanLoadDuringTests(t *testing.T) {
	o := Options{}
	o.Test()
	ls := ":8080"
	if o.Listen != ls {
		t.Fatalf("expected %q got %q", ls, o.Listen)
	}
}
