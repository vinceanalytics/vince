package main

import (
	"flag"
	"net/http"

	"github.com/gernest/len64/app"
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(app.Public))
	svr := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	svr.ListenAndServe()
}
