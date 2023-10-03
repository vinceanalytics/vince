package functions

import (
	"fmt"
	"io"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/vinceanalytics/vince/internal/engine/block"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/scopes"
	"gopkg.in/src-d/go-errors.v1"
)

var ErrInvalidNonLiteralArgument = errors.NewKind("Invalid argument to %s: %s – only literal values supported")

type BaseStats struct {
	site sql.Expression
	db   sql.Database
}

var _ sql.TableFunction = (*BaseStats)(nil)

var baseStatsSSchema = sql.Schema{
	&sql.Column{Name: "timestamp", Type: types.Timestamp},
	&sql.Column{Name: "page_views", Type: types.Int64},
	&sql.Column{Name: "visitors", Type: types.Int64},
	&sql.Column{Name: "visits", Type: types.Int64},
	&sql.Column{Name: "bounce_rate", Type: types.Float64},
	&sql.Column{Name: "visit_duration", Type: types.Float64},
}

func (b *BaseStats) NewInstance(ctx *sql.Context, db sql.Database, expressions []sql.Expression) (sql.Node, error) {
	o := &BaseStats{db: db}
	return o.WithExpressions(expressions...)
}

func (b *BaseStats) Database() sql.Database { return b.db }
func (b *BaseStats) Name() string           { return "base_stats" }
func (b *BaseStats) String() string         { return b.Name() }
func (b *BaseStats) IsReadOnly() bool       { return true }
func (b *BaseStats) Schema() sql.Schema     { return baseStatsSSchema }
func (b *BaseStats) Children() []sql.Node   { return nil }

func (b *BaseStats) Resolved() bool {
	return b.site != nil && b.site.Resolved()
}

func (b *BaseStats) WithDatabase(database sql.Database) (sql.Node, error) {
	o := *b
	o.db = database
	return &o, nil
}

func (b *BaseStats) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, fmt.Errorf("unexpected children")
	}
	return b, nil
}

func (b *BaseStats) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	if b.site == nil {
		return false
	}
	if !types.IsText(b.site.Type()) {
		return false
	}
	return session.Get(ctx).Allow(scopes.GetSite) == nil
}

func (b *BaseStats) Expressions() []sql.Expression {
	o := []sql.Expression{}
	if b.site != nil {
		o = append(o, b.site)
	}
	return o
}

func (b *BaseStats) WithExpressions(expression ...sql.Expression) (sql.Node, error) {
	if len(expression) < 1 {
		return nil, sql.ErrInvalidArgumentNumber.New(b.Name(), "1 to 3", len(expression))
	}
	for _, expr := range expression {
		if !expr.Resolved() {
			return nil, ErrInvalidNonLiteralArgument.New(b.Name(), expr.String())
		}
		// prepared statements resolve functions beforehand, so above check fails
		if _, ok := expr.(sql.FunctionExpression); ok {
			return nil, ErrInvalidNonLiteralArgument.New(b.Name(), expr.String())
		}
	}
	o := *b
	o.site = expression[0]
	return &o, nil
}

func (b *BaseStats) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	domain, err := b.site.Eval(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &statsRowIter{
		MetaIter: block.NewMetaIter(
			session.Get(ctx).DB(), domain.(string),
		),
	}, nil
}

type statsRowIter struct {
	*block.MetaIter
}

var _ sql.RowIter = (*statsRowIter)(nil)

func (s *statsRowIter) Next(_ *sql.Context) (sql.Row, error) {
	if !s.MetaIter.Next() {
		return nil, io.EOF
	}
	b, err := s.Block()
	if err != nil {
		return nil, err
	}
	stat := b.Stats
	o := make(sql.Row, 6)
	o[0] = b.CreatedAt.AsTime()
	o[1] = stat.PageViews
	o[2] = stat.Visitors
	o[3] = stat.Visits
	o[4] = stat.BounceRate
	o[5] = stat.Duration
	return o, nil
}
