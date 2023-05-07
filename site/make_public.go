package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
)

func MakePublic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	models.ChangeSiteVisibility(ctx, site, true)
	session, r := sessions.Load(r)
	session.SuccessFlash(fmt.Sprintf("Stats for %s are now public.", site.Domain))
	session.Save(ctx, w)
	to := fmt.Sprintf("/%s/settings", models.SafeDomain(site))
	http.Redirect(w, r, to, http.StatusFound)
}
