package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
)

func SaveSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := models.GetUser(ctx)
	u.FullName = r.Form.Get("full_name")
	models.Get(ctx).Save(u)
	http.Redirect(w, r, "/settings#profile", http.StatusFound)
}
