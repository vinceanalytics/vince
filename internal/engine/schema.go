package engine

import (
	"bytes"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	vdb "github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
)

var Columns, Indexed = func() (o []storev1.Column, idx map[string]bool) {
	idx = make(map[string]bool)
	for i := storev1.Column_bounce; i <= storev1.Column_utm_term; i++ {
		if i > storev1.Column_timestamp {
			idx[i.String()] = true
		}
		o = append(o, i)
	}
	return
}()

// Creates a schema for a site table. Each site is treated as an individual read
// only table.
//
// Physically timestamps are stored as int64, but we expose this a DateTime.
func Schema(table string, columns []storev1.Column) (o sql.Schema) {
	for _, i := range columns {
		if i <= storev1.Column_timestamp {
			if i == storev1.Column_timestamp {
				o = append(o, &sql.Column{
					Name:     i.String(),
					Type:     types.Timestamp,
					Nullable: false,
					Source:   table,
				})
				continue
			}
			if i == storev1.Column_duration {
				o = append(o, &sql.Column{
					Name:     i.String(),
					Type:     types.Float64,
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
	db.DB.Txn(false, func(txn vdb.Txn) error {
		key := keys.Site(tblName)
		defer key.Release()
		if txn.Has(key.Bytes()) {
			table = &Table{Context: db.Context,
				name:   tblName,
				schema: Schema(tblName, Columns)}
			ok = true
		}
		return nil
	})
	return
}

func (db *DB) GetTableNames(ctx *sql.Context) (names []string, err error) {
	db.DB.Txn(false, func(txn vdb.Txn) error {
		key := keys.Site("")
		it := txn.Iter(vdb.IterOpts{
			Prefix: key.Bytes(),
		})
		for it.Rewind(); it.Valid(); it.Next() {
			names = append(names,
				string(bytes.TrimPrefix(it.Key(), key.Bytes())))
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
