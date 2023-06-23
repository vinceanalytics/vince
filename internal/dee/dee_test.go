package dee

import (
	"bytes"
	"html/template"
	"os"
	"testing"
	"time"

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

const timeseries = `
{{- 
chart | series .ts .values |render
	 -}}
`

func TestSeries(t *testing.T) {
	m := make(template.FuncMap)
	for k, v := range Map {
		m[k] = v
	}
	x := template.Must(template.New("some").Funcs(m).Parse(timeseries))
	base, err := time.Parse(time.RFC822, time.RFC822)
	if err != nil {
		t.Fatal(err)
	}

	values := []float64{1.0, 1.2, 1.3, 1.4, 1.5}
	ts := make([]time.Time, len(values))
	for i := range ts {
		ts[i] = base.AddDate(0, 0, i)
	}
	var b bytes.Buffer
	err = x.Execute(&b, map[string]any{
		"ts":     ts,
		"values": values,
	})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := os.ReadFile("testdata/ts.svg")
	got := b.Bytes()
	if !bytes.Equal(want, got) {
		t.Error("timeseries chart changed")
	}
}
