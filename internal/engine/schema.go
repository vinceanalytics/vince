package engine

import (
	"bytes"

	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/keys"
	v1 "github.com/vinceanalytics/vince/proto/v1"
)

var Columns = func() (o []v1.Column) {
	for i := v1.Column_bounce; i <= v1.Column_utm_term; i++ {
		o = append(o, i)
	}
	return
}()

// Creates a schema for a site table. Each site is treated as an individual read
// only table.
//
// Physically timestamps are stored as int64, but we expose this a DateTime.
func Schema(table string, columns []v1.Column) (o sql.Schema) {
	for _, i := range columns {
		if i <= v1.Column_timestamp {
			if i == v1.Column_timestamp {
				o = append(o, &sql.Column{
					Name:     i.String(),
					Type:     types.Timestamp,
					Nullable: false,
					Source:   table,
				})
				continue
			}
			o = append(o, &sql.Column{
				Name:     i.String(),
				Type:     types.Int64,
				Nullable: false,
				Source:   table,
			})
			continue
		}
		o = append(o, &sql.Column{
			Name:     i.String(),
			Type:     types.Text,
			Nullable: false,
			Source:   table,
		})
	}
	return
}

type DB struct {
	Context
}

var _ sql.Database = (*DB)(nil)

func (DB) Name() string {
	return "vince"
}

func (db *DB) GetTableInsensitive(ctx *sql.Context, tblName string) (table sql.Table, ok bool, err error) {
	db.DB.View(func(txn *badger.Txn) error {
		key := keys.Site(tblName)
		_, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		table = &Table{Context: db.Context,
			name:   tblName,
			schema: Schema(tblName, Columns)}
		ok = true
		return nil
	})
	return
}

func (db *DB) GetTableNames(ctx *sql.Context) (names []string, err error) {
	db.DB.View(func(txn *badger.Txn) error {
		prefix := keys.Site("") + "/"
		o := badger.DefaultIteratorOptions
		o.PrefetchValues = false
		o.Prefix = []byte(prefix)
		it := txn.NewIterator(o)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			names = append(names, string(bytes.TrimPrefix(it.Item().Key(), o.Prefix)))
		}
		return nil
	})
	return
}

func (DB) IsReadOnly() bool {
	return true
}

var _ sql.DatabaseProvider = (*Provider)(nil)

type Provider struct {
	Context
}

func (p *Provider) Database(_ *sql.Context, name string) (sql.Database, error) {
	if name != "vince" {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return &DB{Context: p.Context}, nil
}

func (p *Provider) AllDatabases(_ *sql.Context) []sql.Database {
	return []sql.Database{&DB{Context: p.Context}}
}

func (p *Provider) HasDatabase(_ *sql.Context, name string) bool {
	return name == "vince"
}
