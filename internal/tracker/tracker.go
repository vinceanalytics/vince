package tracker

import (
	"bytes"
	_ "embed"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"text/template"

	"github.com/chanced/powerset"
	"github.com/tdewolff/minify/v2/js"
	"github.com/vinceanalytics/vince/internal/must"
)

//go:embed track.tmpl
var data string

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

var setMapping sync.Map
var cached sync.Map
var once sync.Once
var minifier js.Minifier

func Serve(w http.ResponseWriter, r *http.Request) {
	once.Do(generate)
	file := strings.TrimPrefix(r.URL.Path, "/js/")
	// First use cached
	if b, ok := cached.Load(file); ok {
		w.Write(b.([]byte))
		return
	}
	if b, ok := setMapping.Load(file); ok {
		var buf bytes.Buffer
		must.One(Render(&buf, b.(map[string]bool)))(
			"failed to render track template ", file,
		)
		var mini bytes.Buffer
		must.One(minifier.Minify(nil, &mini, &buf, nil))(
			"failed to minify track file", file,
		)
		cached.Store(file, mini.Bytes())
		w.Write(mini.Bytes())
		return
	}
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

func generate() {
	sets := powerset.Compute([]string{"hash", "outbound-links", "exclusions", "local", "manual", "file-downloads", "dimensions", "tagged-events"})
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
		setMapping.Store(file, m)
	}
	setMapping.Store("vince.js", map[string]bool{})
}
