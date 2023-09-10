package tokens

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"google.golang.org/grpc/credentials"
)

func Valid(priv ed25519.PrivateKey, token string) (ok bool) {
	_, ok = ValidWithClaims(priv, token)
	return
}

func ValidWithClaims(priv ed25519.PrivateKey, token string) (claims *jwt.RegisteredClaims, ok bool) {
	if token == "" {
		return
	}
	claims = &jwt.RegisteredClaims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return priv.Public(), nil
	})
	return claims, err == nil && t.Valid
}

type tokenKey struct{}

type accountKey struct{}

func Set(ctx context.Context, claims *jwt.RegisteredClaims) context.Context {
	return context.WithValue(ctx, tokenKey{}, claims)
}

func Get(ctx context.Context) *jwt.RegisteredClaims {
	return ctx.Value(tokenKey{}).(*jwt.RegisteredClaims)
}

func SetAccount(ctx context.Context, a *v1.Account) context.Context {
	return context.WithValue(ctx, accountKey{}, a)
}

func GetAccount(ctx context.Context) *v1.Account {
	return ctx.Value(accountKey{}).(*v1.Account)
}

type Source string

var _ credentials.PerRPCCredentials = (*Source)(nil)

func (s Source) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + string(s),
	}, nil
}

func (Source) RequireTransportSecurity() bool {
	return false
}

type Basic struct {
	Username string
	Password string
}

var _ credentials.PerRPCCredentials = (*Basic)(nil)

func (b Basic) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	auth := b.Username + ":" + b.Password
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	return map[string]string{
		"authorization": "Basic " + enc,
	}, nil
}

func (Basic) RequireTransportSecurity() bool {
	return false
}
