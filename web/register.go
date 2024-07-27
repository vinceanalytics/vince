package web

import "net/http"

func RegisterForm(w http.ResponseWriter, r *http.Request) {
	register.Execute(w, map[string]any{})
}
