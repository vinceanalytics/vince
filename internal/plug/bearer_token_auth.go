package plug

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/render"
	"github.com/vinceanalytics/vince/internal/tokens"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func AuthGRPC(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		return nil, err
	}
	claims, ok := tokens.ValidWithClaims(db.Get(ctx), token)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid auth token")
	}
	ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", claims.Subject})
	return tokens.Set(ctx, claims), nil
}
