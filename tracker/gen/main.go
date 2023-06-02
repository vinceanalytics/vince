package main

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/chanced/powerset"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/vinceanalytics/vince/tools"
)

//go:embed track.tmpl
var data []byte

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

func main() {
	sets := powerset.Compute([]string{"hash", "outbound-links", "exclusions", "local", "manual", "file-downloads", "dimensions", "tagged-events"})
	var b bytes.Buffer
	var files []string
	os.RemoveAll("out")
	os.MkdirAll("out", 0755)
	for _, v := range sets {
		if len(v) == 0 {
			continue
		}
		sort.Strings(v)
		m := make(map[string]bool)
		for _, x := range v {
			m[strings.Replace(x, "-", "_", -1)] = true
		}
		b.Reset()
		file := strings.Join(append([]string{"vince"}, v...), ".") + ".js"
		err := tpl.Execute(&b, m)
		if err != nil {
			tools.Exit(file, err.Error())
		}
		out := filepath.Join("out", file)
		files = append(files, out)
		tools.WriteFile(out, b.Bytes())
	}
	b.Reset()
	file := "vince.js"
	err := tpl.Execute(&b, map[string]bool{})
	if err != nil {
		tools.Exit(file, err.Error())
	}
	out := filepath.Join("out", file)
	files = append(files, out)
	tools.WriteFile(out, b.Bytes())
	result := api.Build(api.BuildOptions{
		EntryPoints: files,
		Outdir: filepath.Join(
			tools.RootVince(), "assets", "js",
		),
		Target:            api.ES2016,
		Outbase:           "out",
		Write:             true,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
	})
	if len(result.Errors) > 0 {
		e := make([]string, len(result.Errors))
		for i := range e {
			e[i] = result.Errors[i].Text
		}
		tools.Exit(e...)
	}
}
