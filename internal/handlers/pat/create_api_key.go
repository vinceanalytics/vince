package pat

import (
	"net/http"
	"strconv"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/sessions"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/schema"
	"github.com/vinceanalytics/vince/pkg/secrets"
)

func Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	usr := models.GetUser(ctx)
	key := secrets.APIKey()
	name := r.Form.Get("personal_access_token_name")
	expires := r.Form.Get("personal_access_token_expires_at")
	i, _ := strconv.Atoi(expires)
	scopes, err := schema.ParseScopes(r.Form["personal_access_token_scope"]...)
	if err != nil {
		log.Get().Err(err).Msg("failed to parse scopes")
	}

	models.CreatePersonalAccessToken(ctx, key, name, usr.ID, i, scopes)
	session, r := sessions.Load(r)
	session.CustomFlash(key).
		SuccessFlash("API key created successfully").Save(ctx, w)
	http.Redirect(w, r, "/settings#tokens-list", http.StatusFound)
}
