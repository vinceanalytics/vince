package site

import (
	"fmt"
	"net/http"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
)

func InviteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)

	usr := models.UserByEmail(ctx, r.Form.Get("email"))
	if usr != nil && models.UserIsMember(ctx, usr.ID, site.ID) {
		r = sessions.SaveCsrf(w, r)
		render.HTML(ctx, w, templates.InviteMemberForm, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			ctx.Form = r.Form
			ctx.Errors["failed"] = fmt.Sprintf(
				"Cannot send invite because %s is already a member of %s",
				usr.Email, site.Domain,
			)
		})
		return
	}
	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
