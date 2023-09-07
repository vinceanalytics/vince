package output

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/query/v1"
	"github.com/vinceanalytics/vince/internal/px"
)

func Tab(out io.Writer, result *v1.QueryResponse) error {
	w := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)
	var s strings.Builder
	for i, col := range result.Columns {
		if i != 0 {
			s.WriteByte('\t')
		}
		s.WriteString(col.Name)
		s.WriteByte(' ')
	}
	fmt.Fprintln(w, s.String())
	for _, row := range result.Rows {
		s.Reset()
		for i, v := range row.Values {
			if i != 0 {
				s.WriteByte('\t')
			}
			formatValue(&s, v)
		}
		s.WriteByte('\t')
		fmt.Fprintln(w, s.String())
	}
	return w.Flush()
}

func formatValue(w io.Writer, v *v1.QueryValue) {
	fmt.Fprint(w, px.Interface(v))
}
