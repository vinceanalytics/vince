package auth

import (
	"net/http"

	"github.com/gernest/vince/models"
	"github.com/gernest/vince/params"
	"github.com/gernest/vince/sessions"
)

func DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a := models.APIKeyByID(ctx, params.Get(ctx)["id"])
	if a != nil {
		err := models.Get(ctx).Delete(a).Error
		if err != nil {
			models.LOG(ctx, err, "failed to delete api key")
		}
	}
	session, r := sessions.Load(r)
	session.SuccessFlash("API key revoked successfully").Save(ctx, w)
	http.Redirect(w, r, "/settings#api-keys", http.StatusFound)
}
