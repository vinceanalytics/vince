package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/models"
	"github.com/vinceanalytics/vince/sessions"
)

func CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	key := r.Form.Get("key")
	name := r.Form.Get("name")
	models.CreateApiKey(ctx, key, name, usr.ID)
	session, r := sessions.Load(r)
	session.SuccessFlash("API key created successfully").Save(ctx, w)
	http.Redirect(w, r, "/settings#api-keys", http.StatusFound)
}
