package site

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/templates"
)

func InviteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	site := models.GetSite(ctx)
	user := models.GetUser(ctx)
	email := r.Form.Get("email")
	role := r.Form.Get("role")
	usr := models.UserByEmail(ctx, email)
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
	invite := &models.Invitation{
		Email:  email,
		Role:   role,
		SiteID: site.ID,
		UserID: user.ID,
	}
	err := models.Get(ctx).Save(invite).Error
	if err != nil {
		session, r := sessions.Load(r)

		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			session.FailFlash(
				"This invitation has been already sent. To send again, remove it from pending invitations first.",
			)

		} else {
			session.FailFlash("something went wrong")
			models.LOG(ctx, err, "failed to save invite")
		}
		session.Save(ctx, w)
		to := fmt.Sprintf("/%s/settings", models.SafeDomain(site))
		http.Redirect(w, r, to, http.StatusFound)
		return
	}

	render.ERROR(r.Context(), w, http.StatusNotImplemented)
}
