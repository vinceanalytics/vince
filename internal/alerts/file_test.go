package alerts

import (
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	a, err := Create(`
	let g =1;
	__schedule__(100,()=>{g+=1;});
	`)
	if err != nil {
		t.Fatal(err)
	}
	ms := time.Millisecond * 100
	u, ok := a.calls[ms]
	if !ok {
		t.Fatal("expected a scheduled function")
	}
	u.Call()
	got := a.runtime.Get("g").ToInteger()
	want := int64(2)
	if got != want {
		t.Errorf("expected %d got %d", want, got)
	}
}
