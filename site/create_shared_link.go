package site

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/models"
)

func CreateSharedLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	name := r.Form.Get("name")
	password := r.Form.Get("password")
	models.CreateSharedLink(ctx, site.ID, name, password)
	to := fmt.Sprintf("/%s/settings#visibility", models.SafeDomain(site))
	http.Redirect(w, r, to, http.StatusFound)
}
