package auth

import (
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
	// User model is preloaded with verification codes
	for _, codes := range usr.EmailVerificationCodes {
		if codes.Code == uint64(code) {
			// we found a match
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
			txn := models.Get(ctx).Begin()
			// update user  email_verified field
			err := txn.Model(usr).Update("email_verified", true).Error
			if err != nil {
				txn.Rollback()
				log.Get(ctx).Err(err).Msg("failed to update user.email_verified")
				render.ERROR(ctx, w, http.StatusInternalServerError)
				return
			}
			err = txn.Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID).Update("user_id", nil).Error
			if err != nil {
				txn.Rollback()
				log.Get(ctx).Err(err).Msg("failed to update email_verification_codes.user_id")
				render.ERROR(ctx, w, http.StatusInternalServerError)
				return
			}
			err = txn.Commit().Error
			if err != nil {
				txn.Rollback()
				log.Get(ctx).Err(err).Msg("failed to commit email_verification_codes transaction")
				render.ERROR(ctx, w, http.StatusInternalServerError)
				return
			}
			if hasInvitation {
				http.Redirect(w, r, "/sites", http.StatusFound)
				return
			}
			http.Redirect(w, r, "/sites/new", http.StatusFound)
			return
		}

	}
	r = sessions.SaveCsrf(w, r)
	render.HTML(r.Context(), w, templates.Activate, http.StatusOK, func(ctx *templates.Context) {
		ctx.Errors = map[string]string{
			"code": "Incorrect activation code",
		}
		ctx.HasPin = true
		ctx.HasInvitation = hasInvitation
	})
}
