package vince

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
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
			r = r.WithContext(context.WithValue(r.Context(), csrfTokenCtxKey{}, template.HTML(
				fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
					csrfTokenKey, value)),
			))
			h.ServeHTTP(w, r)
			return
		}
		token := make([]byte, 32)
		rand.Read(token)
		session.Data[csrfTokenKey] = string(token)
		cookie := session.Save(w)
		w.Header().Set("Vary", "Cookie")
		r = r.WithContext(context.WithValue(r.Context(), csrfTokenCtxKey{}, template.HTML(
			fmt.Sprintf(`<input type="hidden" name="%s" value="%s">`,
				csrfTokenKey, cookie)),
		))
		h.ServeHTTP(w, r)
	})
}
