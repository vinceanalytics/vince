package a2

import (
	"encoding/base64"

	"github.com/google/uuid"
)

// AuthorizeTokenGenDefault is the default authorization token generator
type AuthorizeTokenGenDefault struct {
}

// GenerateAuthorizeToken generates a base64-encoded UUID code
func (a *AuthorizeTokenGenDefault) GenerateAuthorizeToken(data *AuthorizeData) (ret string, err error) {
	token := uuid.New()
	return base64.RawURLEncoding.EncodeToString(token[:]), nil
}

// AccessTokenGenDefault is the default authorization token generator
type AccessTokenGenDefault struct {
}

// GenerateAccessToken generates base64-encoded UUID access and refresh tokens
func (a *AccessTokenGenDefault) GenerateAccessToken(data *AccessData, generaterefresh bool) (accesstoken string, refreshtoken string, err error) {
	token := uuid.New()
	accesstoken = base64.RawURLEncoding.EncodeToString(token[:])

	if generaterefresh {
		rtoken := uuid.New()
		refreshtoken = base64.RawURLEncoding.EncodeToString(rtoken[:])
	}
	return
}
