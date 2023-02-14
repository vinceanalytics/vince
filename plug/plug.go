package plug

import (
	"net/http"
	"time"

	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
)

type Plug func(http.Handler) http.Handler

func Chain(h http.Handler, plugs ...Plug) http.Handler {
	for i := range plugs {
		h = plugs[len(plugs)-1-i](h)
	}
	return h
}

func Browser() []Plug {
	return []Plug{
		FetchSession,
		PutSecureBrowserHeaders,
		SessionTimeout,
		Auth,
		LastSeen,
	}
}

func SecureForm() []Plug {
	return []Plug{
		Captcha,
		CSRF,
	}
}

func FetchSession(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, r = sessions.Load(r)
		h.ServeHTTP(w, r)
	})
}

func SessionTimeout(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		now := time.Now()
		switch {
		case session.Data.CurrentUserID != 0 && !session.Data.TimeoutAt.IsZero() && now.After(session.Data.TimeoutAt):
			session.Data = sessions.Data{}
		case session.Data.CurrentUserID != 0:
			session.Data.TimeoutAt = now.Add(24 * 7 * 2 * time.Hour)
			session.Save(w)
		}
		h.ServeHTTP(w, r)
	})
}

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		if session.Data.CurrentUserID != 0 {
			usr := &models.User{}
			if err := models.Get(r.Context()).First(usr, session.Data.CurrentUserID).Error; err != nil {
				log.Get(r.Context()).Err(err).Msg("failed fetching current user")
			} else {
				r = r.WithContext(models.SetCurrentUser(r.Context(), usr))
			}
		}
		h.ServeHTTP(w, r)
	})
}

func LastSeen(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		usr := models.GetCurrentUser(r.Context())
		now := time.Now()
		switch {
		case usr != nil && !session.Data.LastSeen.IsZero() && now.Add(-4*time.Hour).After(session.Data.LastSeen):
			usr.LastSeen = now
			err := models.Get(r.Context()).Model(usr).Update("last_seen", now).Error
			if err != nil {
				log.Get(r.Context()).Err(err).Msg("failed to update last_seen")
			}
			session.Data.LastSeen = now
			session.Save(w)
		case usr != nil && session.Data.LastSeen.IsZero():
			session.Data.LastSeen = now
			session.Save(w)
		}
		h.ServeHTTP(w, r)
	})
}

func PutSecureBrowserHeaders(h http.Handler) http.Handler {
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

func Captcha(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
			r = sessions.SaveCaptcha(w, r)
		default:
		}
		h.ServeHTTP(w, r)
	})
}

func CSRF(h http.Handler) http.Handler {
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
