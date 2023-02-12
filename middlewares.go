package vince

import (
	"net/http"
	"time"

	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
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
		_, r = sessions.Load(r)
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) sessionTimeout(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
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
		session, r := sessions.Load(r)
		if userId, ok := session.Data[models.CurrentUserID]; ok {
			usr := &models.User{}
			if err := v.sql.First(usr, uint64(userId.(int64))).Error; err != nil {
				log.Get(r.Context()).Err(err).Msg("failed fetching current user")
			} else {
				r = r.WithContext(models.SetCurrentUser(r.Context(), usr))
			}
		}
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) lastSeen(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
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
				log.Get(r.Context()).Err(err).Msg("failed to update last_seen")
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
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			r = sessions.SaveCaptcha(w, r)
		default:
		}
		h.ServeHTTP(w, r)
	})
}

func (v *Vince) csrf(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			if !sessions.IsValidCSRF(r) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			h.ServeHTTP(w, r)
			return
		}
		r = sessions.SaveCsrf(w, r)
		h.ServeHTTP(w, r)
	})
}
