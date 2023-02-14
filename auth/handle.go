package auth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/email"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/render"
	"github.com/gernest/vince/sessions"
	"gorm.io/gorm"
)

func LoginForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.Login, http.StatusOK)
}

func RegisterForm(w http.ResponseWriter, r *http.Request) {
	render.HTML(r.Context(), w, templates.Register, http.StatusOK)
}

func Register(w http.ResponseWriter, r *http.Request) {
	session, r := sessions.Load(r)
	r.ParseForm()
	u, m, err := NewUser(r)
	if err != nil {
		log.Get(r.Context()).Err(err).Msg("Failed decoding new user from")
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}

	validCaptcha := session.VerifyCaptchaSolution(r)
	if len(m) > 0 || !validCaptcha {
		r = sessions.SaveCsrf(w, r)
		r = sessions.SaveCaptcha(w, r)
		if !validCaptcha {
			if m == nil {
				m = make(map[string]string)
			}
			m["_captcha"] = "Please complete the captcha to register"
		}
		render.HTML(r.Context(), w, templates.Register, http.StatusOK, func(ctx *templates.Context) {
			ctx.Errors = m
			ctx.Form = r.Form
		})
		return
	}
	if err := models.Get(r.Context()).Save(u).Error; err != nil {
		log.Get(r.Context()).Err(err).Msg("failed saving new user")
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			r = sessions.SaveCsrf(w, r)
			r = sessions.SaveCaptcha(w, r)
			render.HTML(r.Context(), w, templates.Register, http.StatusOK, func(ctx *templates.Context) {
				ctx.Errors = map[string]string{
					"email": "already exists",
				}
				ctx.Form = r.Form
			})
			return
		}
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}
	ctx := models.SetCurrentUser(r.Context(), u)
	session.Data.CurrentUserID = u.ID
	session.Data.LoggedIn = true
	session.Save(w)
	if u.EmailVerified {
		http.Redirect(w, r, "/new", http.StatusFound)
	} else {
		err := SendVerificationEmail(ctx, u)
		if err != nil {
			log.Get(r.Context()).Err(err).Msg("failed sending email message")
		}
		http.Redirect(w, r, "/activate", http.StatusFound)
	}
}

func SendVerificationEmail(ctx context.Context, usr *models.User) error {
	code, err := IssueEmailVerification(ctx, usr)
	if err != nil {
		return err
	}
	ctx = templates.SetActivationCode(ctx, code)
	return email.SendActivation(ctx)
}

func ActivateForm(w http.ResponseWriter, r *http.Request) {
	usr := models.GetCurrentUser(r.Context())
	var count int64
	err := models.Get(r.Context()).Model(&models.Invitation{}).Where("email=?", usr.Email).Count(&count).Error
	if err != nil {
		log.Get(r.Context()).Err(err).Msg("failed querying invitation")
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}
	hasInvitation := count != 0
	var code models.EmailVerificationCode
	err = models.Get(r.Context()).Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID).First(&code).Error
	if err != nil {
		log.Get(r.Context()).Err(err).Msg("failed querying invitation")
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}
	ctx := templates.SetActivationCode(r.Context(), code.Code)
	render.HTML(ctx, w, templates.Activate, http.StatusOK, func(ctx *templates.Context) {
		ctx.HasInvitation = hasInvitation
	})
}

func Activate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ctx := r.Context()
	usr := models.GetCurrentUser(ctx)
	var count int64
	err := models.Get(ctx).Model(&models.Invitation{}).Where("email=?", usr.Email).Count(&count).Error
	if err != nil {
		log.Get(r.Context()).Err(err).Msg("failed querying invitation")
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}
	hasInvitation := count != 0
	code, _ := strconv.Atoi(r.Form.Get("code"))
	count = 0
	var codes models.EmailVerificationCode
	err = models.Get(ctx).Where("user_id=?", usr.ID).Where("code=?", code).First(&codes).Error
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
		render.Error(r.Context(), w, http.StatusInternalServerError)
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
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}
	err = models.Get(ctx).Model(&models.EmailVerificationCode{}).Where("user_id=?", usr.ID).Update("user_id", nil).Error
	if err != nil {
		db.Rollback()
		log.Get(r.Context()).Err(err).Msg("failed resetting  verification codes")
		render.Error(r.Context(), w, http.StatusInternalServerError)
		return
	}
	db.Commit()
	if hasInvitation {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/new", http.StatusFound)
}
