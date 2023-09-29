package engine

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/db"
)

type DB struct {
	views map[string]sql.ViewDefinition
}

var _ sql.Database = (*DB)(nil)
var _ sql.ViewDatabase = (*DB)(nil)

func NewDB() *DB {
	return &DB{
		views: make(map[string]sql.ViewDefinition)}
}

func (DB) Name() string {
	return "vince"
}

func (db *DB) GetTableInsensitive(ctx *sql.Context, tblName string) (table sql.Table, ok bool, err error) {
	switch tblName {
	case "sites":
		return &SitesTable{
			// sites table adds name column that returns the site name
			schema: createSchema(append([]string{"name"}, Columns...)),
		}, false, nil
	default:
		return
	}

}

func (DB) GetTableNames(ctx *sql.Context) (names []string, err error) {
	names = append(names, SitesTableName)
	return
}

func (DB) IsReadOnly() bool {
	return true
}

func (db *DB) CreateView(ctx *sql.Context, name string, selectStatement, createViewStmt string) error {
	_, ok := db.views[name]
	if ok {
		return sql.ErrExistingView.New(name)
	}
	sqlMode, _ := sql.LoadSqlMode(ctx)
	db.views[name] = sql.ViewDefinition{
		Name:                name,
		TextDefinition:      selectStatement,
		CreateViewStatement: createViewStmt,
		SqlMode:             sqlMode.String(),
	}
	return nil
}

func (db *DB) DropView(ctx *sql.Context, name string) error {
	_, ok := db.views[name]
	if !ok {
		return sql.ErrViewDoesNotExist.New(db.Name(), name)
	}

	delete(db.views, name)
	return nil
}

func (db *DB) GetViewDefinition(ctx *sql.Context, viewName string) (sql.ViewDefinition, bool, error) {
	viewDef, ok := db.views[viewName]
	return viewDef, ok, nil
}

func (db *DB) AllViews(ctx *sql.Context) ([]sql.ViewDefinition, error) {
	var views []sql.ViewDefinition
	for _, def := range db.views {
		views = append(views, def)
	}
	return views, nil
}

var _ sql.DatabaseProvider = (*Provider)(nil)
var _ sql.FunctionProvider = (*Provider)(nil)

type Provider struct {
	db     db.Provider
	reader b3.Reader
}

func (p *Provider) Function(ctx *sql.Context, name string) (sql.Function, error) {
	fn, ok := funcs[strings.ToLower(name)]
	if !ok {
		return nil, sql.ErrFunctionNotFound.New(name)
	}
	return fn, nil
}

func (p *Provider) Database(_ *sql.Context, name string) (sql.Database, error) {
	if name != "vince" {
		return nil, sql.ErrDatabaseNotFound.New(name)
	}
	return NewDB(), nil
}

func (p *Provider) AllDatabases(_ *sql.Context) []sql.Database {
	return []sql.Database{NewDB()}
}

func (p *Provider) HasDatabase(_ *sql.Context, name string) bool {
	return name == "vince"
}
