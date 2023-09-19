package engine

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
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
	cets := make([]sql.ColumnExpressionType, 0, len(idx.Exprs))
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

func (idx *Index) rangeFilterExpr(ctx *sql.Context, ranges ...sql.Range) (sql.Expression, *Filters, error) {
	exprs := idx.Exprs
	var rangeCollectionExpr sql.Expression
	for rangIdx, rang := range ranges {
		if len(exprs) < len(rang) {
			return nil, nil, fmt.Errorf("expected different key count: exprs(%d) < (ranges[%d])(%d)", len(exprs), rangIdx, len(rang))
		}
		var rangeExpr sql.Expression
		for i, rce := range rang {
			createFilter(exprs[i], rce)
			var rangeColumnExpr sql.Expression
			switch rce.Type() {
			// Both Empty and All may seem like strange inclusions, but if only one range is given we need some
			// expression to evaluate, otherwise our expression would be a nil expression which would panic.
			case sql.RangeType_Empty:
				rangeColumnExpr = expression.NewEquals(expression.NewLiteral(1, types.Int8), expression.NewLiteral(2, types.Int8))
			case sql.RangeType_All:
				rangeColumnExpr = expression.NewEquals(expression.NewLiteral(1, types.Int8), expression.NewLiteral(1, types.Int8))
			case sql.RangeType_EqualNull:
				rangeColumnExpr = expression.NewIsNull(exprs[i])
			case sql.RangeType_GreaterThan:
				if sql.RangeCutIsBinding(rce.LowerBound) {
					rangeColumnExpr = expression.NewGreaterThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.LowerBound), rce.Typ.Promote()))
				} else {
					rangeColumnExpr = expression.NewNot(expression.NewIsNull(exprs[i]))
				}
			case sql.RangeType_GreaterOrEqual:
				rangeColumnExpr = expression.NewGreaterThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.LowerBound), rce.Typ.Promote()))
			case sql.RangeType_LessThanOrNull:
				rangeColumnExpr = expression.JoinOr(
					expression.NewLessThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote())),
					expression.NewIsNull(exprs[i]),
				)
			case sql.RangeType_LessOrEqualOrNull:
				rangeColumnExpr = expression.JoinOr(
					expression.NewLessThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote())),
					expression.NewIsNull(exprs[i]),
				)
			case sql.RangeType_ClosedClosed:
				rangeColumnExpr = expression.JoinAnd(
					expression.NewGreaterThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.LowerBound), rce.Typ.Promote())),
					expression.NewLessThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote())),
				)
			case sql.RangeType_OpenOpen:
				if sql.RangeCutIsBinding(rce.LowerBound) {
					rangeColumnExpr = expression.JoinAnd(
						expression.NewGreaterThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.LowerBound), rce.Typ.Promote())),
						expression.NewLessThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote())),
					)
				} else {
					// Lower bound is (NULL, ...)
					rangeColumnExpr = expression.NewLessThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote()))
				}
			case sql.RangeType_OpenClosed:
				if sql.RangeCutIsBinding(rce.LowerBound) {
					rangeColumnExpr = expression.JoinAnd(
						expression.NewGreaterThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.LowerBound), rce.Typ.Promote())),
						expression.NewLessThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote())),
					)
				} else {
					// Lower bound is (NULL, ...]
					rangeColumnExpr = expression.NewLessThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote()))
				}
			case sql.RangeType_ClosedOpen:
				rangeColumnExpr = expression.JoinAnd(
					expression.NewGreaterThanOrEqual(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.LowerBound), rce.Typ.Promote())),
					expression.NewLessThan(exprs[i], expression.NewLiteral(sql.GetRangeCutKey(rce.UpperBound), rce.Typ.Promote())),
				)
			}
			rangeExpr = expression.JoinAnd(rangeExpr, rangeColumnExpr)
		}
		rangeCollectionExpr = expression.JoinOr(rangeCollectionExpr, rangeExpr)
	}
	re, err := expression.NewRangeFilterExpr(idx.Exprs, ranges)
	if err != nil {
		return nil, nil, err
	}
	return re, &Filters{}, nil
}

func createFilter(field sql.Expression, rce sql.RangeColumnExpr) (i IndexFilter, v ValueFilter) {
	f, ok := field.(*expression.GetField)
	if !ok {
		return
	}
	if _, ok := Indexed[f.Name()]; !ok {
		return
	}
	switch rce.Type() {
	// Both Empty and All may seem like strange inclusions, but if only one range is given we need some
	// expression to evaluate, otherwise our expression would be a nil expression which would panic.
	case sql.RangeType_Empty:
	case sql.RangeType_All:
	case sql.RangeType_EqualNull:
	case sql.RangeType_GreaterThan:
	case sql.RangeType_GreaterOrEqual:
	case sql.RangeType_LessThanOrNull:
	case sql.RangeType_LessOrEqualOrNull:
	case sql.RangeType_ClosedClosed:
	case sql.RangeType_OpenOpen:
	case sql.RangeType_OpenClosed:
	case sql.RangeType_ClosedOpen:
	}
	return
}
