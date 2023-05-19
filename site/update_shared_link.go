package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/sessions"
)

func UpdateSharedLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	name := r.Form.Get("name")
	password := r.Form.Get("password")
	slug := params.Get(ctx)["slug"]
	shared := models.GetSharedLinkWithSlug(ctx, site.ID, slug)
	to := fmt.Sprintf("/%s/settings#visibility", models.SafeDomain(site))
	if shared == nil {
		session, r := sessions.Load(r)
		session.FailFlash("Shared link does not exist").Save(ctx, w)
		http.Redirect(w, r, to, http.StatusFound)
		return
	}
	models.UpdateSharedLink(ctx, shared, name, password)
	http.Redirect(w, r, to, http.StatusFound)
}
