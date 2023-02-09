package vince

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
)

const csrfTokenKey = "_csrf"

func (v *Vince) csrf(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			r.ParseForm()
			value := r.Form.Get(csrfTokenKey)
			saved, ok := session.Data[csrfTokenKey]
			if !ok || subtle.ConstantTimeCompare([]byte(value), []byte(saved.(string))) == 0 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			h.ServeHTTP(w, r)
			return
		}
		token := make([]byte, 32)
		rand.Read(token)
		session.Data[csrfTokenKey] = base64.StdEncoding.EncodeToString(token)
		h.ServeHTTP(w, saveCsrf(session, w, r))
	})
}

func saveCsrf(session *SessionContext, w http.ResponseWriter, r *http.Request) *http.Request {
	solution, data, err := createCaptcha()
	if err != nil {
		xlg.Fatal().Msgf("failed to generate captcha %v", err)
	}
	session.Data[captchaKey] = formatCaptchaSolution(solution)
	session.Save(w)
	w.Header().Set("Vary", "Cookie")
	return r.WithContext(
		templates.SecureForm(r.Context(),
			template.HTML(session.Data[csrfTokenKey].(string)),
			template.HTML(fmt.Sprintf(`<img src="%s" class="img-responsive"/>`, string(data))),
		),
	)
}

func formatCaptchaSolution(sol []byte) string {
	var s strings.Builder
	s.Grow(len(sol))
	for _, b := range sol {
		s.WriteString(strconv.FormatInt(int64(b), 10))
	}
	return s.String()
}
