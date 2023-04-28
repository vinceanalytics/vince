package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
)

func MakePrivate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	models.ChangeSiteVisibility(ctx, site, false)
	session, r := sessions.Load(r)
	session.SuccessFlash(fmt.Sprintf("Stats for %s are now private.", site.Domain))
	session.Save(w)
	to := fmt.Sprintf("/%s/settings/visibility", models.SafeDomain(site))
	http.Redirect(w, r, to, http.StatusFound)
}
