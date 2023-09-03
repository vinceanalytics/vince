package plug

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
	"github.com/vinceanalytics/vince/internal/tokens"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func AuthGRPCBasicSelector(_ context.Context, c interceptors.CallMeta) bool {
	return c.FullMethod() == "/v1.Vince/Login"
}

func AuthGRPCBasic(ctx context.Context) (context.Context, error) {
	token, err := auth.AuthFromMD(ctx, "basic")
	if err != nil {
		return nil, err
	}
	username, passsword, ok := parseBasicAuth(token)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid basic auth")
	}
	var a v1.Account
	err = db.Get(ctx).Txn(false, func(txn db.Txn) error {
		key := keys.Account(username)
		defer key.Release()
		return txn.Get(key.Bytes(), px.Decode(&a), func() error {
			return status.Error(codes.Unauthenticated, "invalid basic auth")
		})
	})
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword(a.HashedPassword, []byte(passsword))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid basic auth")
	}
	ctx = logging.InjectFields(ctx, logging.Fields{"auth.sub", a.Name})
	return tokens.SetAccount(ctx, &a), nil
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	c, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}
