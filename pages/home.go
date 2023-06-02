package pages

import (
	"net/http"

	"github.com/vinceanalytics/vince/models"
)

func Home(w http.ResponseWriter, r *http.Request) {
	cts := r.Context()
	if models.GetUser(cts) != nil {
		http.Redirect(w, r, "/sites", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
