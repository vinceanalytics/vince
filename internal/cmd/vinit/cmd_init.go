package vinit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/dlclark/regexp2"
	"github.com/gympass/goprompt"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initializes a vince project",
		Flags: []cli.Flag{
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
				EnvVars: []string{"VINCE_PASSWORD"},
			},
		},
		Action: func(ctx *cli.Context) error {
			Step("configure root account")

			var name, passwd string
			interactive := ctx.Bool("i")
			if interactive {
				name, passwd = showPrompt()
			} else {
				name = ctx.String("username")
				passwd = ctx.String("password")
			}
			if !ctx.IsSet("i") {
				name, passwd = showPrompt()
			}
			must.Assert(name != "" && passwd != "")(
				"missing root username or password",
			)
			hashed := must.Must(bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost))(
				"failed hashing root password",
			)
			account := must.Must(proto.Marshal(&v1.Account{
				Name:     name,
				Password: hashed,
			}))(
				"failed encoding root account",
			)
			accountKey := keys.Account(name)
			root := ctx.Args().First()
			if root == "" {
				//Default root to current working directory
				root = "."
			}
			if root != "" && root != "." {
				// Try to make sure the directory exists. No need to check for error here
				// because we bail out first thing when we cant write to this path.
				os.MkdirAll(root, 0755)
			}
			Step("using root :%s", root)

			meta := config.META_PATH
			if root != "" {
				meta = filepath.Join(root, meta)
			}

			must.One(os.Mkdir(meta, 0755))(
				"failed to create metadata directory",
			)
			Ok("metadata path :%s", meta)
			blocks := config.BLOCKS_PATH
			if root != "" {
				blocks = filepath.Join(root, blocks)
			}
			must.One(os.Mkdir(blocks, 0755))(
				"failed to create blocks directory",
			)
			Ok("blocks path :%s", blocks)

			b := must.Must(pj.MarshalIndent(config.Defaults()))()
			file := config.FILE
			if root != "" {
				file = filepath.Join(root, file)
			}

			must.One(os.WriteFile(file, b, 0600))(
				"failed to create vince configuration",
			)
			Ok("config path :%s", file)

			_, db := db.Open(context.Background(), meta)
			defer db.Close()
			must.One(db.Update(func(txn *badger.Txn) error {
				return txn.Set([]byte(accountKey), account)
			}))(
				"failed saving account",
			)
			fmt.Fprintf(os.Stdout, "\n %s Done \nnext step:\n  %s\n",
				green(zap), black(fmt.Sprintf("vince serve %s", root)),
			)
			return nil
		},
	}
}

var re = regexp2.MustCompile(
	`^[a-z\d](?:[a-z\d]|-(?=[a-z\d])){0,38}$`,
	regexp2.ECMAScript,
)

func showPrompt() (name, passwd string) {
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

const esc = "\033["

func green(s string) string {
	return fmt.Sprintf("%s32m%v%s", esc, s, resetCode)
}

func black(s string) string {
	return fmt.Sprintf("%s30m%v%s", esc, s, resetCode)
}

const (
	glyphCheck     = "✓"
	glyphSelection = "→"
	zap            = "⚡"
)

var resetCode = fmt.Sprintf("%s%dm", esc, 0)

func Ok(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, " %s %s\n", green(glyphCheck), fmt.Sprintf(msg, args...))
}

func Step(msg string, args ...any) {
	fmt.Fprintf(os.Stdout, "%s %s\n", black(glyphSelection), fmt.Sprintf(msg, args...))
}
