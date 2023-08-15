package engine

import (
	"github.com/apache/arrow/go/v13/arrow"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/engine/core"
	"github.com/vinceanalytics/vince/internal/entry"
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

func Table(name string) *core.Table {
	return core.NewTable(name, sql.NewPrimaryKeySchema(Schema(name)), nil)
}

type RadOnly struct {
	*core.Database
}

func (r RadOnly) IsReadOnly() bool {
	return true
}

var _ sql.DatabaseProvider = (*Provider)(nil)

type Provider struct {
	base *RadOnly
}

func (p *Provider) Database(_ *sql.Context, name string) (sql.Database, error) {
	if name != p.base.Name() {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return p.base, nil
}

func (p *Provider) AllDatabases(_ *sql.Context) []sql.Database {
	return []sql.Database{p.base}
}

func (p *Provider) HasDatabase(_ *sql.Context, name string) bool {
	return p.base.Name() == name
}

func Database(name string) *RadOnly {
	return &RadOnly{
		Database: core.NewDatabase(name),
	}
}
