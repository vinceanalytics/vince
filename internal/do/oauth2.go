package do

import (
	"context"

	"github.com/vinceanalytics/vince/internal/scopes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func Endpoint(uri string) oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:   uri + "/authorize",
		TokenURL:  uri + "/token",
		AuthStyle: oauth2.AuthStyleInHeader,
	}
}

func LoginBase(ctx context.Context, endpoint, username, password string) (*oauth2.Token, error) {
	o := oauth2.Config{
		ClientID:     username,
		ClientSecret: password,
		Endpoint:     Endpoint(endpoint),
	}
	return o.PasswordCredentialsToken(ctx, username, password)
}

func Source(endpoint, id, secret string, scope ...scopes.Scope) *clientcredentials.Config {
	var s []string
	if len(scope) == 0 {
		s = append(s, scopes.All.String())
	} else {
		s = make([]string, len(scope))
		for i := range scope {
			s[i] = scope[i].String()
		}
	}
	return &clientcredentials.Config{
		ClientID:     id,
		ClientSecret: secret,
		TokenURL:     endpoint + "/token",
		Scopes:       s,
		AuthStyle:    oauth2.AuthStyleInHeader,
	}
}
