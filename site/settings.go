package site

import (
	"net/http"
)

func Settings(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.URL.Path+"/general", http.StatusFound)
}
