package site

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/caches"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	u := models.GetUser(ctx)

	// remove site from database
	models.DeleteSite(ctx, u, site)

	// remove site from cache
	caches.Site(ctx).Del(site.Domain)

	session, r := sessions.Load(r)
	session.Success("site and site stats have been permanently deleted").Save(ctx, w)
	http.Redirect(w, r, "/"+u.Name, http.StatusFound)
}
