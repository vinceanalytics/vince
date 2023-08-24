package engine

import (
	"bytes"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
)

// Creates a schema for a site table. Each site is treated as an individual read
// only table.
//
// Physically timestamps are stored as int64, but we expose this a DateTime.
func Schema(table string) (o sql.Schema) {
	for i := range entry.All {
		f := &entry.All[i]
		switch f.Name {
		case "timestamp":
			o = append(o, &sql.Column{
				Name:     f.Name,
				Type:     types.Datetime,
				Nullable: false,
				Source:   table,
			})
		default:
			switch f.Type.ID() {
			case arrow.INT64:
				o = append(o, &sql.Column{
					Name:     f.Name,
					Type:     types.Int64,
					Nullable: false,
					Source:   table,
				})
			case arrow.STRING:
				o = append(o, &sql.Column{
					Name:     f.Name,
					Type:     types.Text,
					Nullable: false,
					Source:   table,
				})
			default:
				must.Assert(false)("unsupported field type", f.Type.ID())
			}
		}
	}
	return
}

type DB struct {
	db *badger.DB
}

var _ sql.Database = (*DB)(nil)

func (DB) Name() string {
	return "vince"
}

func (db *DB) GetTableInsensitive(ctx *sql.Context, tblName string) (table sql.Table, ok bool, err error) {
	db.db.View(func(txn *badger.Txn) error {
		key := keys.Site(tblName)
		_, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		table = &Table{name: tblName, schema: Schema(tblName)}
		ok = true
		return nil
	})
	return
}

func (db *DB) GetTableNames(ctx *sql.Context) (names []string, err error) {
	db.db.View(func(txn *badger.Txn) error {
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
	db *badger.DB
}

func (p *Provider) Database(_ *sql.Context, name string) (sql.Database, error) {
	if name != "vince" {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return &DB{db: p.db}, nil
}

func (p *Provider) AllDatabases(_ *sql.Context) []sql.Database {
	return []sql.Database{&DB{db: p.db}}
}

func (p *Provider) HasDatabase(_ *sql.Context, name string) bool {
	return name == "vince"
}
