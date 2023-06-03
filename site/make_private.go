package site

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/sessions"
)

func MakePrivate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	models.ChangeSiteVisibility(ctx, site, false)
	session, r := sessions.Load(r)
	session.SuccessFlash(fmt.Sprintf("Stats for %s are now private.", site.Domain))
	session.Save(ctx, w)
	to := fmt.Sprintf("/%s/settings", models.SafeDomain(site))
	http.Redirect(w, r, to, http.StatusFound)
}
