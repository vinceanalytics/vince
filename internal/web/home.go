package web

import (
	"encoding/json"
	"net/http"

	"github.com/vinceanalytics/vince/internal/web/db"
)

func Home(db *db.Config, w http.ResponseWriter, r *http.Request) {
	if db.CurrentUser() != nil {
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}
	db.HTML(w, home, nil)
}

func Json(w http.ResponseWriter, data any, code int) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
