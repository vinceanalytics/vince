package vince

import (
	"encoding/json"
	"net/http"
)

func ServeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}
