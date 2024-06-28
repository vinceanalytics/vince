package db

import (
	"bytes"
	"fmt"
)

type ViewFmt struct {
	b bytes.Buffer
}

func (v *ViewFmt) Format(view, field string) string {
	v.b.Reset()
	fmt.Fprintf(&v.b, "~%s;%s<", field, view)
	return v.b.String()
}
