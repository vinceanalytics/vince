package vince

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
)

func CSRF() func(w http.ResponseWriter, r *http.Request) bool {
	ss := NewSession("_csrf")
	headerName := "X-CSRF-Token"
	return func(w http.ResponseWriter, r *http.Request) bool {
		session := ss.Load(r)
		fmt.Printf("%#v\n", session.Data)
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			old, ok := session.Data["token"]
			if !ok {
				// the request didn't have the csrf token we reject it
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
			value := r.Header.Get(headerName)
			current := session.s.Get(value)
			if current == nil {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
			data := map[string]any{}
			err := json.Unmarshal(current, &data)
			if err != nil {
				xlg.Err(err).Msg("failed to decode cookie value")
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
			activeToken, ok := data["token"]
			if !ok {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
			os, ok1 := old.(string)
			as, ok2 := activeToken.(string)
			if !ok1 || !ok2 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
			if subtle.ConstantTimeCompare([]byte(os), []byte(as)) == 0 {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return false
			}
			return true
		}
		token := make([]byte, 32)
		rand.Read(token)
		session.Data["token"] = string(token)
		cookie := session.Save(w)
		w.Header().Set("Vary", "Cookie")
		w.Header().Set(headerName, cookie)
		return true
	}
}
