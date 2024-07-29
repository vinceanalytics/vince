package web

import (
	"encoding/json"
	"net/http"

	"github.com/gernest/len64/web/db"
)

func Home(db *db.Config, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		home.Execute(w, map[string]any{})
		return
	}
	e404.Execute(w, db.Context(make(map[string]any)))
}

func Json(w http.ResponseWriter, data any, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
