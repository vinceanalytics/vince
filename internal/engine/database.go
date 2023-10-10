package engine

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/vinceanalytics/vince/internal/engine/functions"
	"github.com/vinceanalytics/vince/internal/engine/procedures"
)

type DB struct {
	views map[string]sql.ViewDefinition
}

var _ sql.Database = (*DB)(nil)
var _ sql.ViewDatabase = (*DB)(nil)

func NewDB() *DB {
	return &DB{
		views: make(map[string]sql.ViewDefinition),
	}
}

func (DB) Name() string {
	return "vince"
}

func (db *DB) GetTableInsensitive(ctx *sql.Context, tblName string) (table sql.Table, ok bool, err error) {
	switch strings.ToLower(tblName) {
	case eventsTableName:
		return &eventsTable{
			schema: createSchema(Columns),
		}, false, nil
	default:
		return
	}

}

func (DB) GetTableNames(ctx *sql.Context) (names []string, err error) {
	names = append(names, eventsTableName)
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
var _ sql.TableFunctionProvider = (*Provider)(nil)
var _ sql.ExternalStoredProcedureProvider = (*Provider)(nil)

type Provider struct {
	externalProcedures sql.ExternalStoredProcedureRegistry
	funcs              map[string]sql.Function
}

func NewProvider() *Provider {
	externalProcedures := sql.NewExternalStoredProcedureRegistry()
	for _, esp := range procedures.Procedures {
		externalProcedures.Register(esp)
	}
	funcs := make(map[string]sql.Function)
	for _, f := range VinceFuncs {
		funcs[strings.ToLower(f.FunctionName())] = f
	}
	return &Provider{externalProcedures: externalProcedures, funcs: funcs}
}

func (p *Provider) Function(ctx *sql.Context, name string) (sql.Function, error) {
	fn, ok := p.funcs[strings.ToLower(name)]
	if !ok {
		return nil, sql.ErrFunctionNotFound.New(name)
	}
	return fn, nil
}

func (p *Provider) TableFunction(_ *sql.Context, name string) (sql.TableFunction, error) {
	switch strings.ToLower(name) {
	case "base_stats":
		return &functions.BaseStats{}, nil
	default:
		return nil, sql.ErrTableFunctionNotFound.New(name)
	}
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

// ExternalStoredProcedure implements the sql.ExternalStoredProcedureProvider interface
func (p *Provider) ExternalStoredProcedure(_ *sql.Context, name string, numOfParams int) (*sql.ExternalStoredProcedureDetails, error) {
	return p.externalProcedures.LookupByNameAndParamCount(name, numOfParams)
}

// ExternalStoredProcedures implements the sql.ExternalStoredProcedureProvider interface
func (p *Provider) ExternalStoredProcedures(_ *sql.Context, name string) ([]sql.ExternalStoredProcedureDetails, error) {
	return p.externalProcedures.LookupByName(name)
}
