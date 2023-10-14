package vinit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dgraph-io/badger/v4"
	"github.com/urfave/cli/v3"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/api/v1"
	"github.com/vinceanalytics/vince/internal/cmd/ansi"
	"github.com/vinceanalytics/vince/internal/cmd/auth"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	"github.com/vinceanalytics/vince/internal/pj"
	"github.com/vinceanalytics/vince/internal/secrets"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Initializes a vince project",
		Flags: auth.Flags,
		Action: func(ctx *cli.Context) error {
			w := ansi.New()
			w.Step("setting up empty vince project").Flush()
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
				return w.Complete(err)
			}
			os.MkdirAll(root, 0755)
			w.KV("root", root)

			dbPath := config.DB_PATH
			if root != "" {
				dbPath = filepath.Join(root, dbPath)
			}

			must.One(os.Mkdir(dbPath, 0755))(
				"failed to create db directory",
			)
			w.KV("db", dbPath)
			blocks := filepath.Join(root, config.BLOCKS_PATH)
			must.One(os.Mkdir(blocks, 0755))(
				"failed to create blocks directory",
			)
			w.KV("blocks", blocks)
			raft := filepath.Join(root, config.RAFT_PATH)
			must.One(os.Mkdir(raft, 0755))(
				"failed to create raft directory",
			)
			w.KV("raft", raft)
			secret := filepath.Join(root, config.SECRET_KEY)
			must.One(os.WriteFile(secret, []byte(secrets.ED25519()), 0600))("failed creating secret key")
			w.KV("secret_key", secret)

			b := must.Must(pj.MarshalIndent(config.Defaults()))(
				"failed encoding default config",
			)

			file := filepath.Join(root, config.FILE)

			must.One(os.WriteFile(file, b, 0600))(
				"failed to create vince configuration",
			)
			w.KV("config file", file).Flush()
			_, db := db.Open(context.Background(), dbPath, "silent")
			defer db.Close()
			must.One(db.Update(func(txn *badger.Txn) error {
				return txn.Set(accountKey, account)
			}))(
				"failed saving account",
			)
			return w.Suggest(
				fmt.Sprintf("vince serve %s", root),
			).Complete(nil)
		},
	}
}
