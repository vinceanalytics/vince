package assets

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/must"
)

var files = map[string]bool{
	"/favicon.svg": true,
	"/favicon.ico": true,
	"/favicon":     true,
	"/robots.txt":  true,
	"/logo.svg":    true,
	"/index.html":  true,
	"/":            true,
}

//go:embed ui
var static embed.FS

var ui = must.Must(fs.Sub(static, "ui"))("failed getting sub directory")

var FS = http.FileServer(http.FS(ui))

func Match(path string) bool {
	return strings.HasPrefix(path, "/static") ||
		strings.HasPrefix(path, "/vs") ||
		strings.HasPrefix(path, "/min-map") ||
		files[path]
}
