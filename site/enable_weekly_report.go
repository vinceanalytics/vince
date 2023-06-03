package site

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/sessions"
)

func EnableWeeklyReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	usr := models.GetUser(ctx)
	models.EnableWeeklyReport(ctx, site, usr)
	session, r := sessions.Load(r)
	session.SuccessFlash("You will receive an email report every Monday going forward")
	session.Save(ctx, w)
	to := fmt.Sprintf("/%s/settings", models.SafeDomain(site))
	http.Redirect(w, r, to, http.StatusFound)
}
