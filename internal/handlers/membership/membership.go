package membership

import (
	"fmt"
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
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

func Invite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.SiteFor(ctx, models.GetUser(ctx).ID, models.GetSite(ctx).Domain)
	user := models.QueryUserByNameOrEmail(ctx, r.FormValue("email"))
	if user != nil && models.IsMember(ctx, user.ID, site.ID) {
		r = sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, inviteMemberForm, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			ctx.Errors["email"] = fmt.Sprintf(
				"Cannot send invite because %s is already a member of %s",
				user.Email, site.Domain,
			)
			ctx.Form = r.Form
		})
		return
	}
	render.HTML(ctx, w, inviteMemberForm, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
	})
}
