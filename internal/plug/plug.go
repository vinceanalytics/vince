package plug

import (
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/flash"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/pkg/log"
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

func (p Pipeline) And(n ...Plug) Pipeline {
	return append(p, n...)
}

func NOOP(w http.ResponseWriter, r *http.Request) {}

func FetchSession(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, r = sessions.Load(r)
		h.ServeHTTP(w, r)
	})
}

func FetchFlash(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		f := session.Data.Flash
		if f != nil {
			r = r.WithContext(flash.Set(r.Context(), f))
			// save session without the flashes
			session.Data.Flash = nil
			session.Save(r.Context(), w)
		}
		h.ServeHTTP(w, r)
	})
}

func Track() Plug {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/js/vince") {
				w.Header().Set("x-content-type-options", "nosniff")
				w.Header().Set("cross-origin-resource-policy", "cross-origin")
				w.Header().Set("access-control-allow-origin", "*")
				w.Header().Set("cache-control", "public, max-age=86400, must-revalidate")
			}
			h.ServeHTTP(w, r)
		})
	}
}

func SessionTimeout(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		now := core.Now(r.Context())
		switch {
		case session.Data.USER != 0 && !session.Data.TimeoutAt.IsZero() && now.After(session.Data.TimeoutAt):
			session.Data = sessions.Data{}
		case session.Data.USER != 0:
			session.Data.TimeoutAt = now.Add(24 * 7 * 2 * time.Hour)
			session.Save(r.Context(), w)
		}
		h.ServeHTTP(w, r)
	})
}

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		if session.Data.USER != 0 {
			if u := models.UserByUID(r.Context(), session.Data.USER); u != nil {
				r = r.WithContext(models.SetUser(r.Context(), u))
			} else {
				session.Data = sessions.Data{}
				session.Save(r.Context(), w)
			}
		}
		h.ServeHTTP(w, r)
	})
}

func LastSeen(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := sessions.Load(r)
		usr := models.GetUser(r.Context())
		now := core.Now(r.Context())
		switch {
		case usr != nil && !session.Data.LastSeen.IsZero() && now.Add(-4*time.Hour).After(session.Data.LastSeen):
			usr.LastSeen = now
			err := models.Get(r.Context()).Model(usr).Update("last_seen", now).Error
			if err != nil {
				log.Get().Err(err).Msg("failed to update last_seen")
			}
			session.Data.LastSeen = now
			session.Save(r.Context(), w)
		case usr != nil && session.Data.LastSeen.IsZero():
			session.Data.LastSeen = now
			session.Save(r.Context(), w)
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
			r = sessions.SaveCsrf(w, r)
		default:
			if !sessions.IsValidCSRF(r) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(w, r)
	})
}

func CORS(h http.Handler) http.Handler {
	var allowedHeaders http.Header
	var once sync.Once
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cors := config.Get(r.Context())
		once.Do(func() {
			allowedHeaders = make(http.Header)
			for _, v := range cors.Cors.Headers {
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
				if cors.Cors.Origin == "*" {
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
				if cors.Cors.MaxAge > 0 {
					headers.Set("Access-Control-Max-Age", strconv.Itoa(cors.Cors.MaxAge))
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
			if cors.Cors.Origin == "*" {
				headers.Set("Access-Control-Allow-Origin", "*")
			} else {
				headers.Set("Access-Control-Allow-Origin", origin)
			}
			if len(cors.Cors.Expose) > 0 {
				headers.Set("Access-Control-Expose-Headers", strings.Join(cors.Cors.Expose, ", "))
			}
			if cors.Cors.Credentials {
				headers.Set("Access-Control-Allow-Credentials", "true")
			}
		}
		h.ServeHTTP(w, r)
	})
}

func isOriginAllowed(c *config.Options, o string) bool {
	if c.Cors.Origin == "*" {
		return true
	}
	ok, _ := path.Match(c.Cors.Origin, o)
	return ok
}
func isMethodAllowed(c *config.Options, m string) bool {
	for i := range c.Cors.Methods {
		if c.Cors.Methods[i] == m {
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
		lg := log.Get()
		rg := lg.With().Str("request_id", r.Header.Get("x-request-id")).Logger()
		r = r.WithContext(rg.WithContext(r.Context()))
		h.ServeHTTP(w, r)
	})
}
