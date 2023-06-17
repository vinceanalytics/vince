package pages

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
)

func Home(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	if u != nil {
		http.Redirect(w, r, "/"+u.Name, http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}
