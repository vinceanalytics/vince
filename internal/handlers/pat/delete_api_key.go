package pat

import (
	"net/http"
	"strconv"

	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/params"
	"github.com/vinceanalytics/vince/internal/sessions"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	aid, _ := strconv.ParseUint(params.Get(ctx)["id"], 10, 64)
	a := models.APIKeyByID(ctx, aid)
	if a != nil {
		err := models.Get(ctx).Delete(a).Error
		if err != nil {
			models.LOG(ctx, err, "failed to delete api key")
		}
	}
	session, r := sessions.Load(r)
	session.Success("API key revoked successfully").Save(ctx, w)
	http.Redirect(w, r, "/settings#tokens-list", http.StatusFound)
}
