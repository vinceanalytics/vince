package vince

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gernest/vince/models"
)

type middleware func(http.Handler) http.Handler

func (v *Vince) browser() []middleware {
	return []middleware{
		v.fetchSession,
		putSecureBrowserHeaders,
		v.sessionTimeout,
		v.auth,
		v.lastSeen,
	}
}

func (v *Vince) secureForm() []middleware {
	return []middleware{
		v.captcha,
		v.csrf,
	}
}

func (v *Vince) fetchSession(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, r = v.clientSession.Load(r)
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) sessionTimeout(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		var timeoutAt time.Time
		if v, ok := session.Data["session_timeout_at"]; ok {
			timeoutAt = time.Unix(v.(int64), 0)
		}
		_, userID := session.Data[models.CurrentUserID]
		now := time.Now()
		switch {
		case userID && now.After(timeoutAt):
			for k := range session.Data {
				delete(session.Data, k)
			}
		case userID:
			session.Data["session_timeout_at"] = now.Add(24 * 7 * 2 * time.Hour)
			session.Save(w)
		}
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		if userId, ok := session.Data[models.CurrentUserID]; ok {
			usr := &models.User{}
			if err := v.sql.First(usr, uint64(userId.(int64))).Error; err != nil {
				xlg.Err(err).Msg("failed fetching current user")
			} else {
				r = r.WithContext(models.SetCurrentUser(r.Context(), usr))
			}
		}
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) lastSeen(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		usr := models.GetCurrentUser(r.Context())
		var lastSeen time.Time
		if v, ok := session.Data["last_seen"]; ok {
			lastSeen = time.Unix(v.(int64), 0)
		}
		now := time.Now()
		switch {
		case usr != nil && !lastSeen.IsZero() && now.Add(-4*time.Hour).After(lastSeen):
			usr.LastSeen = now
			err := v.sql.Model(usr).Update("last_seen", now).Error
			if err != nil {
				xlg.Err(err).Msg("failed to update last_seen")
			}
			session.Data["last_seen"] = now.Unix()
			session.Save(w)
		case usr != nil && lastSeen.IsZero():
			session.Data["last_seen"] = now.Unix()
			session.Save(w)
		}
		h.ServeHTTP(w, r)
	})
}

func putSecureBrowserHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-frame-options", "SAMEORIGIN")
		w.Header().Set("x-xss-protection", "1; mode=block")
		w.Header().Set("x-content-type-options", "nosniff")
		w.Header().Set("x-download-options", "noopen")
		w.Header().Set("x-permitted-cross-domain-policies", "none")
		w.Header().Set("cross-origin-window-policy", "deny")
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) captcha(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			r = saveCaptcha(w, r, session)
		default:
		}
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) csrf(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			r.ParseForm()
			value := r.Form.Get("_csrf")
			saved, ok := session.Data["_csrf"]
			if !ok || subtle.ConstantTimeCompare([]byte(value), []byte(saved.(string))) == 0 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			h.ServeHTTP(w, r)
			return
		}
		r = saveCsrf(w, r, session)
		h.ServeHTTP(w, r)
	})
}

func saveCsrf(w http.ResponseWriter, r *http.Request, session *SessionContext) *http.Request {
	token := make([]byte, 32)
	rand.Read(token)
	csrf := base64.StdEncoding.EncodeToString(token)
	session.Data["_csrf"] = csrf
	session.Save(w)
	return r.WithContext(templates.SetCSRF(r.Context(), template.HTML(csrf)))
}

func formatCaptchaSolution(sol []byte) string {
	var s strings.Builder
	s.Grow(len(sol))
	for _, b := range sol {
		s.WriteString(strconv.FormatInt(int64(b), 10))
	}
	return s.String()
}
