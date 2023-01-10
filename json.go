package vince

import (
	"encoding/json"
	"net/http"
)

func ServeJSON(w http.ResponseWriter, data any) {
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
