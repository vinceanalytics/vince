package login

import (
	"context"
	"crypto/ed25519"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/api/v1"
	configv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/do"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/tokens"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Authenticate into vince instance",
		Flags: auth.Flags,
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			name, password := auth.Load(ctx)
			uri := ctx.Args().First()
			if uri == "" {
				uri = auth.Instance()
			}
			must.Assert(uri != "")(
				"missing instance argument",
			)
			client, file := auth.LoadClient()
			priv := ed25519.PrivateKey(client.PrivateKey)
			if client.Instance == nil {
				client.Instance = make(map[string]*configv1.Client_Instance)
			}
			if client.ServerId == nil {
				client.ServerId = make(map[string]string)
			}
			if client.Instance[uri] == nil {
				client.Instance[uri] = &configv1.Client_Instance{
					Accounts: make(map[string]*configv1.Client_Auth),
				}
			}
			token, _ := tokens.Generate(
				context.Background(),
				priv,
				v1.Token_CLIENT,
				name,
				time.Now().Add(365*24*time.Hour),
			)

			clientAuth, err := do.Login(context.TODO(),
				uri, name, password, &v1.LoginRequest{
					Token:     token,
					PublicKey: priv.Public().(ed25519.PublicKey),
				},
			)
			if err != nil {
				return w.Complete(err)
			}
			a := clientAuth.Auth
			client.Instance[uri].Accounts[a.Name] = a
			client.ServerId[a.ServerId] = uri
			if client.Active == nil {
				client.Active = &configv1.Client_Active{
					Instance: uri,
					Account:  a.Name,
				}
			}
			must.One(os.WriteFile(file,
				must.Must(pj.MarshalIndent(client))(
					"failed encoding config file",
				),
				0600))(
				"failed writing client config", "path", file,
			)
			ansi.New().Ok("signed in %q", a.ServerId).Flush()
			return nil
		},
	}
}
