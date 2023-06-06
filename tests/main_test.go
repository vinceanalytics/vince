package tests_test

import (
	"testing"

	"github.com/vinceanalytics/vince/tests"
)

func TestDrive(t *testing.T) {
	got := tests.Drive(t, "testdata/hello_world.js")
	want := "hello, world"
	if got != want {
		t.Errorf("expected %q got %q", want, got)
	}
}
