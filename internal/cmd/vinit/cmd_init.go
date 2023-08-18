package vinit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
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
		Flags: auth.Flags,
		Action: func(ctx *cli.Context) error {
			ansi.Step("configure root account")
			name, passwd := auth.Load(ctx)
			hashed := must.Must(bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost))(
				"failed hashing root password",
			)
			account := must.Must(proto.Marshal(&v1.Account{
				Name:           name,
				HashedPassword: hashed,
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
			ansi.Step("using root :%s", root)

			meta := config.META_PATH
			if root != "" {
				meta = filepath.Join(root, meta)
			}

			must.One(os.Mkdir(meta, 0755))(
				"failed to create metadata directory",
			)
			ansi.Ok("metadata path :%s", meta)
			blocks := config.BLOCKS_PATH
			if root != "" {
				blocks = filepath.Join(root, blocks)
			}
			must.One(os.Mkdir(blocks, 0755))(
				"failed to create blocks directory",
			)
			ansi.Ok("blocks path :%s", blocks)

			b := must.Must(pj.MarshalIndent(config.Defaults()))(
				"failed encoding default config",
			)
			file := config.FILE
			if root != "" {
				file = filepath.Join(root, file)
			}

			must.One(os.WriteFile(file, b, 0600))(
				"failed to create vince configuration",
			)
			ansi.Ok("config path :%s", file)

			_, db := db.Open(context.Background(), meta)
			defer db.Close()
			must.One(db.Update(func(txn *badger.Txn) error {
				return txn.Set([]byte(accountKey), account)
			}))(
				"failed saving account",
			)
			fmt.Fprintf(os.Stdout, "\n %s Done \nnext step:\n  %s\n",
				ansi.Green(ansi.Zap), ansi.Black(fmt.Sprintf("vince serve %s", root)),
			)
			return nil
		},
	}
}
