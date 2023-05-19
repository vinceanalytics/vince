package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/sessions"
)

func DeleteSharedLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	slug := params.Get(ctx)["slug"]
	shared := models.GetSharedLinkWithSlug(ctx, site.ID, slug)
	to := fmt.Sprintf("/%s/settings#visibility", models.SafeDomain(site))
	if shared == nil {
		session, r := sessions.Load(r)
		session.FailFlash("Could not find Shared Link").Save(ctx, w)
		http.Redirect(w, r, to, http.StatusFound)
		return
	}
	models.DeleteSharedLink(ctx, shared)
	session, r := sessions.Load(r)
	session.SuccessFlash("Shared Link deleted").Save(ctx, w)
	http.Redirect(w, r, to, http.StatusFound)
}
