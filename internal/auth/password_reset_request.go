package auth

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/email"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/internal/templates"
	"github.com/vinceanalytics/vince/pkg/log"
)

func PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	address := r.Form.Get("email")
	session, r := sessions.Load(r)
	validCaptcha := session.VerifyCaptchaSolution(r)
	if !validCaptcha || address == "" {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		render.HTML(r.Context(), w, templates.PasswordResetRequestForm, http.StatusOK, func(ctx *templates.Context) {
			if !validCaptcha {
				ctx.Errors["captcha"] = "Please complete the captcha to reset your password"
			}
			if address == "" {
				ctx.Errors["email"] = "Please enter an email address"
			}
		})
		return
	}
	ctx := r.Context()
	usr := models.UserByEmail(r.Context(), address)
	if usr != nil {
		token, err := createResetToken(ctx, address)
		if err != nil {
			log.Get().Err(err).Msg("failed to sign token for password reset")
		} else {
			q := make(url.Values)
			q.Set("token", token)
			link := config.Get(ctx).URL + "/password/reset?" + q.Encode()
			err = email.SendPasswordReset(ctx, usr, link)
			if err != nil {
				log.Get().Err(err).Msg("failed to send email verification link")
			}
		}
	}
	render.HTML(ctx, w, templates.PasswordResetRequestSuccess, http.StatusOK, func(ctx *templates.Context) {
		ctx.Email = address
	})
}

func createResetToken(ctx context.Context, email string) (string, error) {
	now := core.Now(ctx)
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, jwt.RegisteredClaims{
		Issuer:    "vince",
		Subject:   "password-reset",
		Audience:  jwt.ClaimStrings{email},
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        uuid.NewString(),
	})
	return token.SignedString(config.GetSecuritySecret(ctx))
}
