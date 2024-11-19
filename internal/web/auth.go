package web

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/vinceanalytics/vince/internal/web/db"
)

func UserSetting(db *db.Config, w http.ResponseWriter, r *http.Request) {
	keys, err := db.Ops().APIKeys()
	if err != nil {
		db.Logger().Error("reading api keys", "err", err)
		db.HTML(w, e500, nil)
		return
	}
	db.HTML(w, userSettings, map[string]any{
		"keys": keys,
	})
}

func NewApiKey(db *db.Config, w http.ResponseWriter, r *http.Request) {
	data := make([]byte, 64)
	rand.Read(data)
	key := base64.StdEncoding.EncodeToString(data)[:64]
	db.HTML(w, newAPIKey, map[string]any{
		"key": key,
	})
}

func DeleteAPiKey(db *db.Config, w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	err := db.Ops().DeleteAPIKey(name)
	if err != nil {
		db.Logger().Error("deleting api key", "name", name, "err", err)
		db.HTML(w, e500, nil)
		return
	}
	db.Success("API key revoked successfully")
	db.SaveSession(w)
	http.Redirect(w, r, "/settings#api-keys", http.StatusFound)
}

func CreateAPiKey(db *db.Config, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := r.FormValue("name")
	key := r.FormValue("key")
	valid := map[string]string{}
	if name == "" {
		valid["error_name"] = "missing"
	}
	if key == "" {
		valid["error_key"] = "missing"
	}
	if key != "" && len(key) < 64 {
		valid["error_key"] = "invalid key"
	}
	err := db.Ops().CreateAPIKey(name, key)
	if err != nil {
		db.Logger().Error("creating new key", "err", err)
		db.HTML(w, e500, nil)
		return
	}
	db.Success("API key created successfully")
	db.SaveSession(w)
	http.Redirect(w, r, "/settings#api-keys", http.StatusFound)
}
