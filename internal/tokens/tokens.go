package tokens

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/scopes"
	"google.golang.org/grpc/credentials"
)

type Claims struct {
	jwt.RegisteredClaims
	Scopes scopes.List
}

func (c *Claims) ValidFor(scope scopes.Scope) bool {
	if c.Scopes.Test(uint(scopes.All)) {
		return true
	}
	return c.Scopes.Test(uint(scope)) &&
		len(c.Audience) > 0 && c.Audience[0] == c.Subject
}

func NewClaims(scopeStr string, claims jwt.RegisteredClaims) (claim *Claims, err error) {
	if scopeStr == "" {
		scopeStr = scopes.All.String()
	}
	scope := strings.Split(scopeStr, ",")
	claim = &Claims{RegisteredClaims: claims}
	var s scopes.Scope
	for i := range scope {
		err := s.Parse(scope[i])
		if err != nil {
			return nil, err
		}
		claim.Scopes.Set(uint(s))
	}
	return
}

func Valid(priv ed25519.PrivateKey, token string, method scopes.Scope) (ok bool) {
	_, ok = ValidWithClaims(priv, token, method)
	return
}

func ValidWithClaims(priv ed25519.PrivateKey, token string, method scopes.Scope) (claims *Claims, ok bool) {
	if token == "" {
		return
	}
	claims = &Claims{}
	t, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return priv.Public(), nil
	})
	return claims, err == nil && t.Valid && claims.ValidFor(method)
}

type tokenKey struct{}

type accountKey struct{}

func Set(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, tokenKey{}, claims)
}

func Get(ctx context.Context) *Claims {
	return ctx.Value(tokenKey{}).(*Claims)
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
