package web

import (
	_ "embed"
	"io"
	"net/http"

	"github.com/vinceanalytics/vince/ref"
)

//go:embed placeholder_favicon.ico
var favicon []byte

func Placeholder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=2592000")
	w.WriteHeader(http.StatusOK)
	w.Write(favicon)
}

func Favicon(w http.ResponseWriter, r *http.Request) {
	source := r.PathValue("source")
	if file, err := ref.Favicon.Open(source); err == nil {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=2592000")
		w.WriteHeader(http.StatusOK)
		io.Copy(w, file)
		return
	}
	Placeholder(w, r)
}
