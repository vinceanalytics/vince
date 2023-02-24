package auth

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
	"gorm.io/gorm"
)

func Activate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ctx := r.Context()
	usr := models.GetCurrentUser(ctx)

	hasInvitation := models.Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Invitation{}).Where("email=?", usr.Email)
	})

	code, _ := strconv.Atoi(r.Form.Get("code"))
	var codes models.EmailVerificationCode
	err := models.Get(ctx).Where("user_id=?", usr.ID).Where("code=?", code).First(&codes).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r = sessions.SaveCsrf(w, r)
			render.HTML(r.Context(), w, templates.Activate, http.StatusOK, func(ctx *templates.Context) {
				ctx.Errors = map[string]string{
					"code": "Incorrect activation code",
				}
				ctx.HasPin = true
				ctx.HasInvitation = hasInvitation
			})
			return
		}
		log.Get(r.Context()).Err(err).Msg("failed querying verification codes")
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	if codes.UpdatedAt.Before(time.Now().Add(-4 * time.Hour)) {
		// verification code has expired
		r = sessions.SaveCsrf(w, r)
		render.HTML(r.Context(), w, templates.Activate, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors = map[string]string{
				"code": "Code is expired, please request another one",
			}
			ctx.HasPin = true
			ctx.HasInvitation = hasInvitation
		})
		return
	}
	// set user.email_verified = true and release all verification codes assigned
	// to user.
	// Do all of this inside a single transaction.
	db := models.Get(ctx).Begin()
	usr.EmailVerified = true
	err = models.Get(ctx).Save(usr).Error
	if err != nil {
		db.Rollback()
		log.Get(r.Context()).Err(err).Msg("failed updating user verification status")
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	err = models.Get(ctx).Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID).Update("user_id", nil).Error
	if err != nil {
		db.Rollback()
		log.Get(r.Context()).Err(err).Msg("failed resetting  verification codes")
		render.ERROR(r.Context(), w, http.StatusInternalServerError)
		return
	}
	db.Commit()
	if hasInvitation {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/new", http.StatusFound)
}
