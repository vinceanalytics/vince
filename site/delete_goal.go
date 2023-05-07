package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/flash"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/sessions"
)

func DeleteGoal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	session, r := sessions.Load(r)
	to := fmt.Sprintf("/%s/settings", models.SafeDomain(site))
	if !models.DeleteGoal(ctx, params.Get(ctx)["id"], site.Domain) {
		session.Data.Flash = &flash.Flash{
			Error: []string{"failed to delete goal"},
		}
	} else {
		session.Data.Flash = &flash.Flash{
			Success: []string{"Goal deleted successfully"},
		}
	}
	session.Save(ctx, w)
	http.Redirect(w, r, to, http.StatusFound)
}
