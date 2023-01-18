package vince

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"net/http"
)

const csrfHeaderName = "X-CSRF-Token"
const csrfTokenKey = "_csrf"

func (v *Vince) csrf(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, r := v.clientSession.Load(r)
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			old, ok := session.Data[csrfTokenKey]
			if !ok {
				// the request didn't have the csrf token we reject it
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			value := r.Header.Get(csrfHeaderName)
			current := session.s.Get(value)
			if current == nil {
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
		session.Data["token"] = string(token)
		cookie := session.Save(w)
		w.Header().Set("Vary", "Cookie")
		w.Header().Set(csrfHeaderName, cookie)
	})
}
