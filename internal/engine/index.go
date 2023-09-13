package engine

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
)

type Index struct {
	DB         string // required for engine tests with driver
	Tbl        *Table // required for engine tests with driver
	TableName  string
	Name       string
	Exprs      []sql.Expression
	exprString []string
	PrefixLens []uint16
}

var _ sql.Index = (*Index)(nil)

func (idx *Index) Database() string                    { return idx.DB }
func (idx *Index) ColumnExpressions() []sql.Expression { return idx.Exprs }
func (idx *Index) IsGenerated() bool                   { return false }
func (idx *Index) Expressions() []string               { return idx.exprString }
func (idx *Index) ExtendedExpressions() []string       { return idx.exprString }
func (idx *Index) CanSupport(...sql.Range) bool        { return true }
func (idx *Index) IsUnique() bool                      { return false }
func (idx *Index) IsSpatial() bool                     { return false }
func (idx *Index) IsFullText() bool                    { return false }
func (idx *Index) Comment() string                     { return "" }
func (idx *Index) IndexType() string                   { return "BTREE" }
func (idx *Index) Table() string                       { return idx.TableName }
func (idx *Index) ID() string                          { return idx.Name }

func (idx *Index) PrefixLengths() []uint16 {
	return idx.PrefixLens
}

func (idx *Index) ColumnExpressionTypes() []sql.ColumnExpressionType {
	cets := make([]sql.ColumnExpressionType, len(idx.Exprs))
	for i, expr := range idx.Exprs {
		cets[i] = sql.ColumnExpressionType{
			Expression: expr.String(),
			Type:       expr.Type(),
		}
	}
	return cets
}

func (idx *Index) ExtendedColumnExpressionTypes() []sql.ColumnExpressionType {
	cets := make([]sql.ColumnExpressionType, 0, len(idx.Tbl.schema))
	cetsInExprs := make(map[string]struct{})
	for _, expr := range idx.Exprs {
		cetsInExprs[strings.ToLower(expr.(*expression.GetField).Name())] = struct{}{}
		cets = append(cets, sql.ColumnExpressionType{
			Expression: expr.String(),
			Type:       expr.Type(),
		})
	}
	return cets
}

func (idx *Index) rangeFilterExpr(ctx *sql.Context, ranges ...sql.Range) (sql.Expression, error) {
	return expression.NewRangeFilterExpr(idx.Exprs, ranges)
}
