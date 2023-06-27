package site

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
)

func Public(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	u := models.GetUser(ctx)
	models.ChangeSiteVisibility(ctx, site, true)
	session, r := sessions.Load(r)
	session.SuccessFlash(fmt.Sprintf("Stats for %s are now public.", site.Domain))
	session.Save(ctx, w)
	to := fmt.Sprintf("/%s/%s/settings#visibility", u.Name, site.Domain)
	http.Redirect(w, r, to, http.StatusFound)
}
