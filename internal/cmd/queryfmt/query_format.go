package queryfmt

import (
	"bytes"
	"io"
	"strings"

	ast "github.com/dolthub/vitess/go/vt/sqlparser"
)

func Format(w io.Writer, query string) error {
	stmt, err := ast.ParseWithOptions(query, ast.ParserOptions{})
	if err != nil {
		return err
	}
	buf := ast.NewTrackedBuffer(pretty())
	stmt.Format(buf)
	w.Write([]byte(buf.String()))
	return nil
}

func pretty() ast.NodeFormatter {
	b := &ast.TrackedBuffer{Builder: new(strings.Builder)}

	var width int

	var prefix bytes.Buffer

	adjust := func() {
		prefix.Reset()
		for n := 0; n < width*2; n++ {
			prefix.WriteByte(' ')
		}
	}
	incr := func() {
		width++
		adjust()
	}
	decr := func() {
		width--
		adjust()
	}
	pad := func(buf *ast.TrackedBuffer) {
		buf.Write(prefix.Bytes())
	}
	add := func(w *ast.TrackedBuffer, node ast.SQLNode) {
		b.Builder.Reset()
		node.Format(b)
		pad(w)
		w.WriteString(b.String())
	}
	return func(buf *ast.TrackedBuffer, node ast.SQLNode) {
		switch e := node.(type) {
		case ast.SelectExprs:
			list(buf, e, add, incr, decr)
			buf.WriteByte('\n')
		case ast.TableExprs:
			list(buf, e, add, incr, decr)
		case ast.GroupBy:
			buf.WriteByte('\n')
			buf.WriteString(" group by ")
			list(buf, e, add, incr, decr)
		case ast.OrderBy:
			list(buf, e, add, incr, decr)
		default:
			add(buf, node)
		}
	}
}

func list[T ast.SQLNode](buf *ast.TrackedBuffer, e []T, add func(*ast.TrackedBuffer, ast.SQLNode), incr, decr func()) {
	buf.WriteByte('\n')
	incr()
	for i, n := range e {
		if i != 0 {
			buf.WriteByte(',')
			buf.WriteByte('\n')
		}
		add(buf, n)
	}
	decr()
}
