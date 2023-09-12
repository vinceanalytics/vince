package a2

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/scopes"
	"github.com/vinceanalytics/vince/internal/secrets"
	"github.com/vinceanalytics/vince/internal/tokens"
)

type JWT struct{}

var _ AccessTokenGen = (*JWT)(nil)
var _ AuthorizeTokenGen = (*JWT)(nil)

func (JWT) GenerateAccessToken(ctx context.Context, data *AccessData, generaterefresh bool) (accesstoken string, refreshtoken string, err error) {
	priv := secrets.Get(ctx)
	var claim *tokens.Claims
	claim, err = tokens.NewClaims(data.Scope, jwt.RegisteredClaims{
		Issuer:  scopes.ResourceBaseURL,
		Subject: data.Client.GetId(),
		ID:      ulid.Make().String(),
		Audience: jwt.ClaimStrings{
			data.Client.GetId(),
		},
		ExpiresAt: jwt.NewNumericDate(data.ExpireAt()),
		IssuedAt:  jwt.NewNumericDate(data.CreatedAt),
	})
	if err != nil {
		return
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claim)

	accesstoken, err = token.SignedString(priv)
	if err != nil {
		return
	}
	if !generaterefresh {
		return
	}
	claim, err = tokens.NewClaims(data.Scope, jwt.RegisteredClaims{
		Issuer:  scopes.ResourceBaseURL,
		Subject: data.Client.GetId(),
		ID:      ulid.Make().String(),
		Audience: jwt.ClaimStrings{
			data.Client.GetId(),
		},
		IssuedAt: jwt.NewNumericDate(data.CreatedAt),
	})
	if err != nil {
		return
	}
	token = jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claim)
	refreshtoken, err = token.SignedString(priv)
	return
}

func (JWT) GenerateAuthorizeToken(ctx context.Context, data *AuthorizeData) (string, error) {
	priv := secrets.Get(ctx)
	claim, err := tokens.NewClaims(data.Scope, jwt.RegisteredClaims{
		Issuer:  scopes.ResourceBaseURL,
		Subject: data.Client.GetId(),
		ID:      ulid.Make().String(),
		Audience: jwt.ClaimStrings{
			data.Client.GetId(),
		},
		ExpiresAt: jwt.NewNumericDate(data.ExpireAt()),
		IssuedAt:  jwt.NewNumericDate(data.CreatedAt),
	})
	if err != nil {
		return "", err
	}
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, claim)
	return token.SignedString(priv)
}
