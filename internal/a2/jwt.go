package a2

import (
	"context"

	"github.com/golang-jwt/jwt/v4"
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
	me := tokens.GetAccount(ctx)
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, &jwt.RegisteredClaims{
		Issuer:  scopes.ResourceBaseURL,
		Subject: me.Name,
		ID:      ulid.Make().String(),
		Audience: jwt.ClaimStrings{
			data.Client.GetId(),
		},
		ExpiresAt: jwt.NewNumericDate(data.ExpireAt()),
		IssuedAt:  jwt.NewNumericDate(data.CreatedAt),
	})

	accesstoken, err = token.SignedString(priv)
	if err != nil {
		return "", "", err
	}
	if !generaterefresh {
		return
	}
	token = jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, &jwt.RegisteredClaims{
		Issuer:  scopes.ResourceBaseURL,
		Subject: me.Name,
		ID:      ulid.Make().String(),
		Audience: jwt.ClaimStrings{
			data.Client.GetId(),
		},
		IssuedAt: jwt.NewNumericDate(data.CreatedAt),
	})
	refreshtoken, err = token.SignedString(priv)
	if err != nil {
		return "", "", err
	}
	return
}

func (JWT) GenerateAuthorizeToken(ctx context.Context, data *AuthorizeData) (string, error) {
	priv := secrets.Get(ctx)
	me := tokens.GetAccount(ctx)
	token := jwt.NewWithClaims(&jwt.SigningMethodEd25519{}, &jwt.RegisteredClaims{
		Issuer:  scopes.ResourceBaseURL,
		Subject: me.Name,
		ID:      ulid.Make().String(),
		Audience: jwt.ClaimStrings{
			data.Client.GetId(),
		},
		ExpiresAt: jwt.NewNumericDate(data.ExpireAt()),
		IssuedAt:  jwt.NewNumericDate(data.CreatedAt),
	})
	return token.SignedString(priv)
}
