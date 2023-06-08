package config

import (
	"testing"
)

func TestCanLoadDuringTests(t *testing.T) {
	o := Test()
	ls := "debug"
	if o.LogLevel != ls {
		t.Fatalf("expected %q got %q", ls, o.Listen)
	}
}
