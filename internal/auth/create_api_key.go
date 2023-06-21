package auth

import (
	"net/http"
	"strconv"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/pkg/secrets"
)

func CreatePersonalAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	key := secrets.APIKey()
	name := r.Form.Get("personal_access_token_name")
	expires := r.Form.Get("personal_access_token_expires_at")
	i, _ := strconv.Atoi(expires)
	models.CreatePersonalAccessToken(ctx, key, name, usr.ID, i)
	session, r := sessions.Load(r)
	session.CustomFlash(key).
		SuccessFlash("API key created successfully").Save(ctx, w)
	http.Redirect(w, r, "/settings#tokens-list", http.StatusFound)
}
