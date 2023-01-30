package vince

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
)

const csrfTokenKey = "_csrf"

type csrfTokenCtxKey struct{}

func getCsrf(ctx context.Context) template.HTML {
	return ctx.Value(csrfTokenCtxKey{}).(template.HTML)
}

type captchaTokenKey struct{}

func getCaptcha(ctx context.Context) template.HTML {
	return ctx.Value(captchaTokenKey{}).(template.HTML)
}

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
			h.ServeHTTP(w, saveCsrf(session, w, r))
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
	session.Data[captchaKey] = base32.StdEncoding.EncodeToString(solution)
	cookie := session.Save(w)
	w.Header().Set("Vary", "Cookie")
	ctx := context.WithValue(r.Context(), csrfTokenCtxKey{}, template.HTML(cookie))
	ctx = context.WithValue(ctx, captchaTokenKey{}, template.HTML(fmt.Sprintf(`<img src="%s" class="img-responsive"/>`, string(data))))
	return r.WithContext(ctx)
}
