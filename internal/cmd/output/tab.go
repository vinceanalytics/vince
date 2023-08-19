package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

func Tab(out io.Writer, cols []*Column) error {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	var s strings.Builder
	for i := range cols {
		if i != 0 {
			s.WriteByte('\t')
		}
		s.WriteString(cols[i].Name)
		s.WriteByte(' ')
	}
	fmt.Fprintln(w, s.String())
	for i := 0; i < len(cols[0].Values); i++ {
		s.Reset()
		for n := range cols {
			if n != 0 {
				s.WriteByte('\t')
			}
			cols[n].Format(&s, i)
		}
		s.WriteByte('\t')
		fmt.Fprintln(w, s.String())
	}
	return w.Flush()
}

type CountWriter struct {
	io.Writer
	N int
}

func (c *CountWriter) Write(p []byte) (n int, err error) {
	n, err = c.Writer.Write(p)
	c.N += n
	return
}
