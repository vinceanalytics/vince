package auth

import (
	"net/http"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/pkg/secrets"
)

func CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	key := secrets.APIKey()
	name := r.Form.Get("personal_api_key_name")
	models.CreateApiKey(ctx, key, name, usr.ID)
	session, r := sessions.Load(r)
	session.CustomFlash(key).
		SuccessFlash("API key created successfully").Save(ctx, w)
	http.Redirect(w, r, "/settings#keys", http.StatusFound)
}
