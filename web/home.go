package web

import (
	"encoding/json"
	"net/http"

	"github.com/gernest/len64/web/db"
)

func Home(db *db.Config, w http.ResponseWriter, r *http.Request) {
	home.Execute(w, map[string]any{})
}

func Json(w http.ResponseWriter, data any, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
