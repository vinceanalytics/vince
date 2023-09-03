package tokens

import (
	"context"
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"google.golang.org/protobuf/proto"
)

func Generate(ctx context.Context, key ed25519.PrivateKey,
	issuer v1.Token_Issuer,
	account string, expires time.Time, scopes ...v1.Token_Scope) (string, *jwt.RegisteredClaims) {
	if len(scopes) == 0 {
		scopes = []v1.Token_Scope{v1.Token_ALL}
	}
	claims := &jwt.RegisteredClaims{
		Issuer:    issuer.String(),
		Subject:   account,
		ID:        ulid.Make().String(),
		ExpiresAt: jwt.NewNumericDate(expires),
		IssuedAt:  jwt.NewNumericDate(core.Now(ctx)),
	}
	for _, scope := range scopes {
		claims.Audience = append(claims.Audience, scope.String())
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claims)
	return must.Must(token.SignedString(key))(
		"failed signing jwt token",
	), claims
}

func Valid(db db.Provider, token string) (ok bool) {
	_, ok = ValidWithClaims(db, token)
	return
}

func ValidWithClaims(vdb db.Provider, token string) (claims *jwt.RegisteredClaims, ok bool) {
	if token == "" {
		return
	}
	ok = vdb.Txn(false, func(txn db.Txn) error {
		key := keys.Token(token)
		defer key.Release()
		return txn.Get(
			key.Bytes(),
			func(val []byte) error {
				var tpub v1.Token
				err := proto.Unmarshal(val, &tpub)
				if err != nil {
					return err
				}
				claims = &jwt.RegisteredClaims{}
				t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
					return ed25519.PublicKey(tpub.PubKey), nil
				})
				if err != nil {
					return err
				}
				if !t.Valid {
					return errors.New("invalid token")
				}
				return nil
			},
		)
	}) == nil

	return
}

type tokenKey struct{}

func Set(ctx context.Context, claims *jwt.RegisteredClaims) context.Context {
	return context.WithValue(ctx, tokenKey{}, claims)
}

func Get(ctx context.Context) *jwt.RegisteredClaims {
	return ctx.Value(tokenKey{}).(*jwt.RegisteredClaims)
}
