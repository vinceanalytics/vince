package dee

import (
	"bytes"
	"html/template"
	"testing"

	"github.com/wcharczuk/go-chart/v2"
)

func TestChain(t *testing.T) {
	m := make(template.FuncMap)
	for k, v := range Map {
		m[k] = v
	}
	m["show"] = func(a *chart.Chart) string {
		return a.Title
	}
	x := template.Must(template.New("some").Funcs(m).Parse(`{{ chart | title "hello, world" | show }}`))
	var b bytes.Buffer
	err := x.Execute(&b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := "hello, world", b.String(); want != got {
		t.Errorf("expected %q got %q", want, got)
	}
}

func TestStyle(t *testing.T) {
	m := make(template.FuncMap)
	for k, v := range Map {
		m[k] = v
	}
	m["show"] = func(a *chart.Chart) string {
		return a.Background.ClassName
	}
	x := template.Must(template.New("some").Funcs(m).Parse(`{{ chart | style "" | class "dee" | show }}`))
	var b bytes.Buffer
	err := x.Execute(&b, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want, got := "dee", b.String(); want != got {
		t.Errorf("expected %q got %q", want, got)
	}
}
