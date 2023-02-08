package vince

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
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
			old, ok := session.Data[csrfTokenKey]
			if !ok {
				// the request didn't have the csrf token we reject it
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			value := r.Form.Get(csrfTokenKey)
			current := session.s.Get(value)
			if current == nil {
				// our cookie value is secure, we failed to decrypt/decode. reject
				// this request
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			data := map[string]any{}
			err := json.Unmarshal(current, &data)
			if err != nil {
				xlg.Err(err).Msg("failed to decode cookie value")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			activeToken, ok := data[csrfTokenKey]
			if !ok {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			os, ok1 := old.(string)
			as, ok2 := activeToken.(string)
			if !ok1 || !ok2 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			if subtle.ConstantTimeCompare([]byte(os), []byte(as)) == 0 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			h.ServeHTTP(w, r)
			return
		}
		token := make([]byte, 32)
		rand.Read(token)
		session.Data[csrfTokenKey] = string(token)
		h.ServeHTTP(w, saveCsrf(session, w, r))
	})
}

func saveCsrf(session *SessionContext, w http.ResponseWriter, r *http.Request) *http.Request {
	solution, data, err := createCaptcha()
	if err != nil {
		xlg.Fatal().Msgf("failed to generate captcha %v", err)
	}
	session.Data[captchaKey] = formatCaptchaSolution(solution)
	cookie := session.Save(w)
	w.Header().Set("Vary", "Cookie")
	return r.WithContext(
		templates.SecureForm(r.Context(),
			template.HTML(cookie),
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
