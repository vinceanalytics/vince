package tokens

import (
	"context"
	"crypto/ed25519"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

func Generate(ctx context.Context, key ed25519.PrivateKey,
	issuer v1.Token_Issuer,
	account string, expires time.Time) (string, *jwt.RegisteredClaims) {
	claims := &jwt.RegisteredClaims{
		Issuer:    issuer.String(),
		Subject:   account,
		ID:        ulid.Make().String(),
		ExpiresAt: jwt.NewNumericDate(expires),
		IssuedAt:  jwt.NewNumericDate(core.Now(ctx)),
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claims)

	return must.Must(token.SignedString(key))(
		"failed signing jwt token",
	), claims

}
