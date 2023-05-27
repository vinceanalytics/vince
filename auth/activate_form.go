package auth

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/templates"
	"gorm.io/gorm"
)

func ActivateForm(w http.ResponseWriter, r *http.Request) {
	usr := models.GetUser(r.Context())
	hasInvitation := models.Exists(r.Context(), func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.Invitation{}).Where("email=?", usr.Email)
	})
	hasCode := models.Exists(r.Context(), func(db *gorm.DB) *gorm.DB {
		return db.Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID)
	})
	render.HTML(r.Context(), w, templates.Activate, http.StatusOK, func(ctx *templates.Context) {
		ctx.HasPin = hasCode
		ctx.HasInvitation = hasInvitation
	})
}
