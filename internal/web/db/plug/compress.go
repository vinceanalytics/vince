package plug

import (
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

const (
	contentEncoding = "Content-Encoding"
)

var pool = &sync.Pool{New: func() any {
	return &wrap{gz: gzip.NewWriter(io.Discard)}
}}

func Compress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/") {
			h.ServeHTTP(w, r)
			return
		}
		wr := pool.Get().(*wrap)
		defer wr.Release()

		wr.Reset(w)
		h.ServeHTTP(wr, r)
		err := wr.Close()
		if err != nil {
			slog.Error("closing gzip stream", "err", err)
		}
	})
}

type wrap struct {
	gz *gzip.Writer
	http.ResponseWriter
}

func (w *wrap) Write(p []byte) (int, error) {
	return w.gz.Write(p)
}

func (w *wrap) Reset(wr http.ResponseWriter) {
	wr.Header().Set(contentEncoding, "gzip")
	w.gz.Reset(wr)
	w.ResponseWriter = wr
}

func (w *wrap) Close() error {
	return w.gz.Close()
}

func (w *wrap) Release() {
	w.gz.Reset(io.Discard)
	w.ResponseWriter = nil
}
