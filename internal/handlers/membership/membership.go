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
	u := models.GetUser(ctx)
	site := models.SiteFor(ctx, u.ID, models.GetSite(ctx).Domain)
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
	invite := &models.Invitation{
		Email:  r.Form.Get("email"),
		Role:   r.Form.Get("role"),
		SiteID: site.ID,
		UserID: u.ID,
	}
	if o := models.NewInvite(ctx, invite); len(o) > 0 {
		r = sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, inviteMemberForm, http.StatusOK, func(ctx *templates.Context) {
			ctx.Site = site
			for k := range o {
				ctx.Errors[k] = o.Get(k)
			}
			ctx.Form = r.Form
		})
		return
	}
	err := models.Get(ctx).Create(invite).Error
	if err != nil {
		models.LOG(ctx, err, "failed to create invitation")
		to := fmt.Sprintf("/%s/%s/settings#people", u.Email, site.Domain)
		session, r := sessions.Load(r)
		session.Fail("something went wrong").Save(ctx, w)
		http.Redirect(w, r, to, http.StatusFound)
		return
	}
	render.HTML(ctx, w, inviteMemberForm, http.StatusOK, func(ctx *templates.Context) {
		ctx.Site = site
	})
}
