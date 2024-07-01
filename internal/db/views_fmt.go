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

func (v *ViewFmt) Shard(view string, shard uint64) string {
	v.b.Reset()
	fmt.Fprintf(&v.b, "%s_%06d", view, shard)
	return v.b.String()
}
