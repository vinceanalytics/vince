package web

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/web/db"
)

func LicenseForm(db *db.Config, w http.ResponseWriter, r *http.Request) {
	db.HTML(w, license, nil)
}

func License(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	key := r.FormValue("license")
	if err := db.Get().ApplyLicense([]byte(key)); err != nil {
		db.SaveCsrf(w)
		db.HTML(w, license, map[string]any{
			"error": err.Error(),
		})
		return
	}
	http.Redirect(w, r, "/sites", http.StatusFound)
}
