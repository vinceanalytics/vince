package js

import (
	"context"
	"testing"
)

func TestImport(t *testing.T) {
	a, err := Compile(context.TODO(), "testdata/schedule.js[spike,15m]")
	if err != nil {
		t.Fatal(err)
	}
	x := a[0]
	if x.Function == nil {
		t.Error("expected schedule function to be called")
	}
	if want, got := "spike", x.Name; want != got {
		t.Errorf("expected %s got %s", want, got)
	}
}
