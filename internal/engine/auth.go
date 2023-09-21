package engine

import (
	"bytes"
	"context"
	"log/slog"

	"github.com/dgraph-io/badger/v4"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/scopes"
	"github.com/vinceanalytics/vince/internal/secrets"
	"github.com/vinceanalytics/vince/internal/tokens"
)

type PlaintextAuthPluginFunc func(db *mysql_db.MySQLDb,
	user string, userEntry *mysql_db.User, pass string) (bool, error)

func (f PlaintextAuthPluginFunc) Authenticate(db *mysql_db.MySQLDb,
	user string, userEntry *mysql_db.User, pass string) (bool, error) {
	return f(db, user, userEntry, pass)
}

var _ mysql_db.PlaintextAuthPlugin = (*PlaintextAuthPluginFunc)(nil)

const authPluginName = "vince"

// sets up authorization of the clients. Clients are forced to use
// mysql_clear_password where password is the jwt access token.
//
// we set the users in e.Analyzer.Catalog.MySQLDb with only
// sql.PrivilegeType_Select privilege. To ensure we always have up to date users
// information we watch key changes on the accounts namespace and refresh the
// users info whenever we detect there were changes in the accounts namespace.
func setupAuth(ctx context.Context, e *sqle.Engine) {
	m := e.Analyzer.Catalog.MySQLDb
	m.SetPlugins(map[string]mysql_db.PlaintextAuthPlugin{
		authPluginName: validateUserAccess(ctx),
	})
	set := func() {
		ed := m.Editor()
		defer ed.Close()
		for _, v := range listUsers(db.Get(ctx)) {
			slog.Debug("adding user to the catalog", "user", v)
			pset := mysql_db.NewPrivilegeSet()
			pset.AddGlobalStatic(
				sql.PrivilegeType_Select,
			)
			ed.PutUser(&mysql_db.User{
				User:         v,
				Plugin:       authPluginName,
				PrivilegeSet: pset,
				Host:         "localhost",
			})
		}
	}
	// ensure we load users when ww startup
	set()
	m.SetEnabled(true)
	prefix := keys.Account("")
	db.Observe(ctx, keys.Account(""), func(ctx context.Context, k *badger.KVList) {
		for i := range k.Kv {
			if bytes.HasPrefix(k.Kv[i].Key, prefix) {
				set()
				break
			}
		}
	})
}

func validateUserAccess(ctx context.Context) PlaintextAuthPluginFunc {
	return func(db *mysql_db.MySQLDb, user string, userEntry *mysql_db.User, pass string) (bool, error) {
		slog.Debug("authorize mysql_db user ", "user", user)
		return tokens.Valid(
			secrets.Get(ctx), pass, scopes.Query), nil
	}
}

func listUsers(provider db.Provider) (usr []string) {
	txn := provider.NewTransaction(false)
	prefix := keys.Account("")
	it := txn.Iter(db.IterOpts{
		Prefix:         prefix,
		PrefetchValues: false,
	})
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		usr = append(usr, string(bytes.TrimPrefix(it.Key(), prefix)))
	}
	return
}
