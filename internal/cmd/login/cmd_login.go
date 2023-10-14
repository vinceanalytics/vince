package login

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"
	configv1 "github.com/vinceanalytics/proto/gen/go/vince/config/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/do"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "Authenticate into vince instance",
		Flags: append(auth.Flags, &cli.BoolFlag{
			Name:  "token",
			Usage: "Prints access token to stdout",
		}),
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
			token, err := do.LoginBase(
				context.TODO(), uri, name, password,
			)
			if err != nil {
				return w.Complete(err)
			}
			info, err := do.Build(context.TODO(), uri, token.AccessToken)
			if err != nil {
				return w.Complete(err)
			}
			a := &configv1.Client_Auth{
				Name:          name,
				AccessToken:   token.AccessToken,
				RerfreshToken: token.RefreshToken,
				ServerId:      info.ServerId,
			}
			client, file := auth.LoadClient()
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
			if ctx.Bool("token") {
				os.Stdout.WriteString(a.AccessToken)
				return nil
			}
			ansi.New().Ok("signed in %q", a.ServerId).Flush()
			return nil
		},
	}
}
