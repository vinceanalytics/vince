package main

import (
	"flag"
	"net/http"

	"github.com/gernest/len64/app"
	"github.com/gernest/len64/web"
)

func main() {
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/public/", http.FileServerFS(app.Public))
	mux.HandleFunc("/", web.Home)
	svr := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	svr.ListenAndServe()
}
