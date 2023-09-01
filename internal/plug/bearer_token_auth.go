package plug

import (
	"net/http"

	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/tokens"
)

func Auth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearer(r.Header)
		ctx := r.Context()
		claims, ok := tokens.ValidWithClaims(db.Get(ctx), token)
		if !ok {
			render.ERROR(w, http.StatusUnauthorized)
			return
		}
		o := config.Get(ctx)
		h.ServeHTTP(w, r.WithContext(core.SetAuth(ctx, &v1.Client_Auth{
			Name:     claims.Subject,
			Token:    token,
			ServerId: o.ServerId,
		})))
	})
}
