package main

import (
	"bytes"
	_ "embed"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/chanced/powerset"
	"github.com/tdewolff/minify/v2/js"
)

//go:embed track.tmpl
var data string

func main() {
	flag.Parse()
	err := generate(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
}

var tpl = template.Must(
	template.New("").Delims("<<", ">>").Funcs(template.FuncMap{
		"hasCustomEvents": func(a map[string]bool) bool {
			for _, v := range []string{"outbound_links", "file_downloads", "tagged_events"} {
				if a[v] {
					return true
				}
			}
			return false
		},
	}).Parse(string(data)),
)

// Render uses feature to generate  tracker js script.
func Render(w io.Writer, features map[string]bool) error {
	return tpl.Execute(w, features)
}

func generate(dir string) error {
	sets := powerset.Compute([]string{"hash", "outbound-links", "exclusions", "local", "manual", "file-downloads", "dimensions", "tagged-events"})
	var b bytes.Buffer

	var minifier js.Minifier
	var o bytes.Buffer
	min := func() []byte {
		o.Reset()
		err := minifier.Minify(nil, &o, &b, nil)
		if err != nil {
			log.Fatal(err)
		}
		return o.Bytes()
	}
	for _, v := range sets {
		if len(v) == 0 {
			continue
		}
		sort.Strings(v)
		m := make(map[string]bool)
		for _, x := range v {
			m[strings.Replace(x, "-", "_", -1)] = true
		}
		file := strings.Join(append([]string{"vince"}, v...), ".") + ".js"
		b.Reset()
		err := Render(&b, m)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(dir, file), min(), 0600)
		if err != nil {
			return err
		}
	}
	b.Reset()
	err := Render(&b, map[string]bool{})
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "vince.js"), min(), 0600)
}
