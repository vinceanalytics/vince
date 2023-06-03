package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
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
