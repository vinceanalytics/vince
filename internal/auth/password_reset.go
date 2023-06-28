package auth

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
)

func PasswordReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	t := r.Form.Get("token")
	password := r.Form.Get("password")
	if password == "" {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, passwordResetFormTpl, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors["password"] = "Password cannot be empty"
			ctx.Form = r.Form
		})
		return
	}
	token, err := jwt.ParseWithClaims(t, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
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
	claims := token.Claims.(*jwt.RegisteredClaims)
	usr := models.QueryUserByNameOrEmail(ctx, claims.Audience[0])
	if usr != nil {
		usr.PasswordHash = models.HashPassword(password)
		err = models.Get(ctx).Save(usr).Error
		if err != nil {
			models.LOG(ctx, err, "failed to update password hash")
			render.ERROR(r.Context(), w, http.StatusInternalServerError)
			return
		}
		session, r := sessions.Load(r)
		session.Success("Password updated successfully").
			Success("Please log in with your new credentials")
		session.Data.LoggedIn = false
		session.Save(ctx, w)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	r = sessions.SaveCsrf(w, r)
	render.HTML(r.Context(), w, passwordResetFormTpl, http.StatusOK, func(ctx *templates.Context) {
		ctx.Token = t
	})
}
