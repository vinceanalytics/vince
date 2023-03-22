package plot

import (
	"bytes"
	"os"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestCreateSVG(t *testing.T) {
	n := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Div,
		Data:     "div",
	}
	renderVerticalGradient(n, "some_id")
	var b bytes.Buffer
	html.Render(&b, n)
	os.WriteFile("out.html", b.Bytes(), 0600)
}
