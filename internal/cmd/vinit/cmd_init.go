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
			root, err := filepath.Abs(ctx.Args().First())
			if err != nil {
				return ansi.ERROR(err)
			}
			os.MkdirAll(root, 0755)
			ansi.Step("using root :%s", root)

			mainDB := config.DB_PATH
			if root != "" {
				mainDB = filepath.Join(root, mainDB)
			}

			must.One(os.Mkdir(mainDB, 0755))(
				"failed to create db directory",
			)
			ansi.Ok("main db path :%s", mainDB)
			blocks := config.BLOCKS_PATH
			if root != "" {
				blocks = filepath.Join(root, blocks)
			}
			must.One(os.Mkdir(blocks, 0755))(
				"failed to create blocks directory",
			)
			ansi.Ok("blocks path :%s", blocks)

			raft := config.RAFT_PATH
			if root != "" {
				raft = filepath.Join(root, raft)
			}
			must.One(os.Mkdir(raft, 0755))(
				"failed to create raft directory",
			)
			ansi.Ok("raft path :%s", raft)

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

			_, db := db.Open(context.Background(), mainDB, "silent")
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
