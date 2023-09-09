package do

import (
	"github.com/vinceanalytics/vince/internal/scopes"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

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
