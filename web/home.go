package web

import (
	"encoding/json"
	"net/http"

	"github.com/gernest/len64/web/db"
)

func Home(db *db.Config, w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		db.HTML(w, home, nil)
		return
	}
	db.HTML(w, e404, nil)
}

func Json(w http.ResponseWriter, data any, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
