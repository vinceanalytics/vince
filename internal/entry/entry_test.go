package entry

import (
	"bytes"
	"testing"
)

func TestWrite(t *testing.T) {
	m := NewMulti()
	var b bytes.Buffer
	w := NewFileWriter(&b)
	m.Append(&Entry{})
	m.Write(w, nil)
	w.Close()

	f := NewFileReader(bytes.NewReader(b.Bytes()))
	r := NewReader()
	r.Read(f, nil)
	x := r.Record()
	if want, got := int64(1), x.NumRows(); want != got {
		t.Errorf("expected %d rows got %d", want, got)
	}
}
