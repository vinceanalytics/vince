package auth

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/templates"
)

var passwordResetFormTpl = templates.Focus("auth/password_reset_form.html")

func PasswordResetForm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := r.URL.Query().Get("token")
	token, err := jwt.Parse(t, func(t *jwt.Token) (interface{}, error) {
		return config.GetSecuritySecret(ctx).Public(), nil
	})
	if err != nil || !token.Valid {
		msg := "Your token is invalid. Please request another password reset link."
		if errors.Is(err, jwt.ErrTokenExpired) {
			msg = "Your token has expired. Please request another password reset link."

		}
		render.ERROR(r.Context(), w, http.StatusUnauthorized, func(ctx *templates.Context) {
			ctx.Error.StatusText = msg
		})
		return
	}
	render.HTML(r.Context(), w, passwordResetFormTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.Token = t
	})
}
