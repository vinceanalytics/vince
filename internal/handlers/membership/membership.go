package membership

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var inviteMemberForm = templates.Focus("site/invite_member_form.html")

func InviteForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	render.HTML(ctx, w, inviteMemberForm, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
	})
}
