package curl

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var m = protojson.MarshalOptions{}

type CMD string

func (cmd CMD) Format(w io.Writer, path string, method string, headers http.Header, body proto.Message) error {
	fmt.Fprintf(w, "curl -X %s", bashEscape(method))
	if body != nil {
		data, err := m.Marshal(body)
		if err != nil {
			return err
		}
		// have data in a new line
		fmt.Fprintf(w, " \\n -d %s", bashEscape(string(data)))
	}
	var keys []string

	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		// have headers in a new line
		fmt.Fprintf(w, " \\n -H %s", bashEscape(fmt.Sprintf("%s: %s", k, strings.Join(headers[k], " "))))
	}
	fmt.Fprintf(w, " %s", bashEscape(string(cmd)+path))
	return nil
}

func bashEscape(str string) string {
	return `'` + strings.Replace(str, `'`, `'\''`, -1) + `'`
}
