package render

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

type err struct {
	Error any `json:"error"`
}

func ERROR(w http.ResponseWriter, code int, msg ...any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if len(msg) == 0 {
		msg = []any{http.StatusText(code)}
	}
	json.NewEncoder(w).Encode(err{Error: msg[0]})
}
