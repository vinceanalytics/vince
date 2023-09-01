package auth

import (
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/dlclark/regexp2"
	"github.com/gympass/goprompt"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/secrets"
)

var re = regexp2.MustCompile(
	`^[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38}$`,
	regexp2.ECMAScript,
)

var Flags = []cli.Flag{
	&cli.BoolWithInverseFlag{
		BoolFlag: &cli.BoolFlag{
			Name:  "i",
			Usage: "Shows interactive prompt for username and password",
		},
	},
	&cli.StringFlag{
		Name:    "username",
		Usage:   "Name of the root user",
		Value:   "root",
		EnvVars: []string{"VINCE_ROOT_USER"},
	},
	&cli.StringFlag{
		Name:    "password",
		Usage:   "password of the root user",
		Value:   "vince",
		EnvVars: []string{"VINCE_ROOT_PASSWORD"},
	},
}

func Load(ctx *cli.Context) (name, passwd string) {
	interactive := ctx.Bool("i")
	if interactive {
		name, passwd = Prompt()
	} else {
		name = ctx.String("username")
		passwd = ctx.String("password")
		if passwd == "stdin" {
			passwd = string(must.Must(io.ReadAll(os.Stdin))(
				"failed reading password from stdin",
			))
		}
	}
	if !ctx.IsSet("i") && (name == "" || passwd == "") {
		name, passwd = Prompt()
	}
	must.Assert(name != "" && passwd != "")(
		"missing root username or password",
	)
	return
}

func Prompt() (name, passwd string) {
	prompt := goprompt.Prompt{
		Label:        "Name",
		DefaultValue: "root",
		Description:  "Root username",
		Validation: func(s string) bool {
			ok, _ := re.MatchString(s)
			return ok
		},
	}
	r := must.Must(prompt.Run())("failed obtaining username")
	must.Assert(!r.Cancelled)("cancelled")
	name = r.Value

	prompt = goprompt.Prompt{
		Label:       "Password",
		Description: "Root password",
		Validation: func(s string) bool {
			return len(s) > 4
		},
	}
	r = must.Must(prompt.Run())("failed obtaining root password")
	must.Assert(!r.Cancelled)("cancelled")
	passwd = r.Value
	return
}

func LoadClient() (client *v1.Client, file string) {
	client = &v1.Client{}
	usr := must.Must(user.Current())(
		"failed getting current user",
	)
	base := filepath.Join(usr.HomeDir, ".vince")

	// we don't check the error here so that we can error out when reading or
	// writing to this directory
	os.MkdirAll(base, 0755)
	file = filepath.Join(base, config.FILE)

	b, err := os.ReadFile(file)
	if err != nil {
		if os.IsNotExist(err) {
			privateKey := secrets.ED25519Raw()
			client.PrivateKey = privateKey
			b := must.Must(
				pj.MarshalIndent(client),
			)(
				"failed marshalling private key",
			)
			must.One(os.WriteFile(file, b, 0600))(
				"failed writing config file", "path", file,
			)
		} else {
			must.One(err)("failed reading config file", "path", file)
		}
	} else {
		must.One(pj.Unmarshal(b, client))(
			"failed decoding client config",
		)
	}
	return
}

func Save(w *ansi.W, client *v1.Client, file string) {
	err := os.WriteFile(file,
		must.Must(pj.MarshalIndent(client))("failed encoding client"),
		0600,
	)
	if err != nil {
		w.Err("failed saving client config path:%q err:%v", file, err).Exit()
	}
}

func Account() (token string, api string) {
	o, _ := LoadClient()
	if o.Active == nil {
		w := ansi.New()
		w.Err("no active account found")
		w.Suggest(
			"log in to a vince instance with [vince login] command",
			"select existing vince instance/account using [vince use] command",
		).Exit()
	}
	token = o.Instance[o.Active.Instance].Accounts[o.Active.Account].Token
	api = o.Active.Instance
	return
}

func Instance() (api string) {
	o, _ := LoadClient()
	if o.Active == nil {
		return
	}
	api = o.Active.Instance
	return
}
