package plug

import (
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/log"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/sessions"
	"github.com/oklog/ulid/v2"
)

type Plug func(http.Handler) http.Handler

type Pipeline []Plug

func (p Pipeline) Pass(h http.HandlerFunc) http.Handler {
	x := http.Handler(h)
	for i := range p {
		x = p[len(p)-1-i](x)
	}
	return x
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

func CORS(h http.Handler) http.Handler {
	var allowedHeaders http.Header
	var once sync.Once
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors := config.Get(r.Context()).Cors
		once.Do(func() {
			allowedHeaders = make(http.Header)
			for _, v := range cors.Headers {
				// we use set as value so we can check if .Get method returns ""
				// will mean the header was not set.
				allowedHeaders.Set(v, "set")
			}
		})
		method := r.Header.Get("access-control-request-method")
		if r.Method == http.MethodOptions && method != "" {
			headers := w.Header()
			origin := r.Header.Get("Origin")

			headers.Add("Vary", "Origin")
			headers.Add("Vary", "Access-Control-Request-Method")
			headers.Add("Vary", "Access-Control-Request-Headers")
			headers.Add("Vary", "Access-Control-Request-Private-Network")
			reqHeaders := parseHeaderList(r.Header.Get("Access-Control-Request-Headers"))
			switch {
			case origin == "":
			case !isMethodAllowed(cors, origin):
			case !isMethodAllowed(cors, r.Header.Get("Access-Control-Request-Method")):
			case !isHeadersAllowed(allowedHeaders, reqHeaders):
			default:
				if cors.Origin == "*" {
					headers.Set("Access-Control-Allow-Origin", "*")
				} else {
					headers.Set("Access-Control-Allow-Origin", origin)
				}
				headers.Set("Access-Control-Allow-Methods", strings.ToUpper(r.Header.Get("Access-Control-Request-Method")))
				if len(reqHeaders) > 0 {
					headers.Set("Access-Control-Allow-Headers", strings.Join(reqHeaders, ", "))
				}
				if r.Header.Get("Access-Control-Request-Private-Network") == "true" {
					headers.Set("Access-Control-Allow-Private-Network", "true")
				}
				if cors.MaxAge > 0 {
					headers.Set("Access-Control-Max-Age", strconv.Itoa(int(cors.MaxAge)))
				}
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		headers := w.Header()
		origin := r.Header.Get("Origin")
		headers.Add("Vary", "Origin")
		switch {
		case origin == "":
		case !isOriginAllowed(cors, origin):
		case !isMethodAllowed(cors, r.Method):
		default:
			if cors.Origin == "*" {
				headers.Set("Access-Control-Allow-Origin", "*")
			} else {
				headers.Set("Access-Control-Allow-Origin", origin)
			}
			if len(cors.Expose) > 0 {
				headers.Set("Access-Control-Expose-Headers", strings.Join(cors.Expose, ", "))
			}
			if cors.Credentials {
				headers.Set("Access-Control-Allow-Credentials", "true")
			}
		}
		h.ServeHTTP(w, r)
	})
}

func isOriginAllowed(c *config.Cors, o string) bool {
	if c.Origin == "*" {
		return true
	}
	ok, _ := path.Match(c.Origin, o)
	return ok
}
func isMethodAllowed(c *config.Cors, m string) bool {
	for i := range c.Methods {
		if c.Methods[i] == m {
			return true
		}
	}
	return false
}

func isHeadersAllowed(h http.Header, o []string) bool {
	for i := range o {
		if h.Get(o[i]) == "" {
			return false
		}
	}
	return true
}

const toLower = 'a' - 'A'

// parseHeaderList tokenize + normalize a string containing a list of headers
func parseHeaderList(headerList string) []string {
	l := len(headerList)
	h := make([]byte, 0, l)
	upper := true
	// Estimate the number headers in order to allocate the right splice size
	t := 0
	for i := 0; i < l; i++ {
		if headerList[i] == ',' {
			t++
		}
	}
	headers := make([]string, 0, t)
	for i := 0; i < l; i++ {
		b := headerList[i]
		switch {
		case b >= 'a' && b <= 'z':
			if upper {
				h = append(h, b-toLower)
			} else {
				h = append(h, b)
			}
		case b >= 'A' && b <= 'Z':
			if !upper {
				h = append(h, b+toLower)
			} else {
				h = append(h, b)
			}
		case b == '-' || b == '_' || b == '.' || (b >= '0' && b <= '9'):
			h = append(h, b)
		}

		if b == ' ' || b == ',' || i == l-1 {
			if len(h) > 0 {
				// Flush the found header
				headers = append(headers, string(h))
				h = h[:0]
				upper = true
			}
		} else {
			upper = b == '-' || b == '_'
		}
	}
	return headers
}

func RequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-request-id") == "" {
			r.Header.Set("x-request-id", ulid.Make().String())
		}
	})
}
