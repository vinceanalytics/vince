package web

import "net/http"

func LoginForm(w http.ResponseWriter, r *http.Request) {
	login.Execute(w, map[string]any{})
}
