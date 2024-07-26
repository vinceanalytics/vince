package web

import "net/http"

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html")
	home.Execute(w, map[string]any{})
}
