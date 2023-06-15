package js

import (
	"context"
	"testing"
)

func TestImport(t *testing.T) {
	file, err := Compile(context.TODO(), "testdata/schedule/")
	if err != nil {
		t.Fatal(err)
	}
	u := file.Units()
	if want, got := 1, len(u); got != want {
		t.Errorf("expected %d got %d", want, got)
	}

	// check that we have reference to the function and we can call it.
	v, err := u[0].(*Unit).calls[0](file.runtime.GlobalObject())
	if err != nil {
		t.Fatal(err)
	}
	if want, got := int64(200), v.ToInteger(); got != want {
		t.Errorf("expected %d got %d", want, got)
	}
}
