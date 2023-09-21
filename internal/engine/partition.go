package engine

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

type IndexHint struct {
	Fields []*FieldHint
}

type FieldHint struct {
	Field string
	Op    Op
	Value any
}

func (v *IndexHint) Visit(node sql.Node, expr sql.Expression) sql.NodeVisitor {
	e, ok := node.(*plan.Filter)
	if !ok {
		return v
	}
	transform.WalkExpressions(&visitFilters{v: v}, e)
	return nil
}

type visitFilters struct {
	v *IndexHint
}

func (v *visitFilters) Visit(expr sql.Expression) sql.Visitor {
	switch e := expr.(type) {
	case *expression.Equals:
		return pickFilter(v, e, Eq)
	case *expression.GreaterThan:
		return pickFilter(v, e, Gt)
	case *expression.GreaterThanOrEqual:
		return pickFilter(v, e, GtEg)
	case *expression.LessThan:
		return pickFilter(v, e, Lt)
	case *expression.LessThanOrEqual:
		return pickFilter(v, e, LtEq)
	default:
		return v
	}
}

type comparison interface {
	sql.Expression
	Left() sql.Expression
	Right() sql.Expression
}

func pickFilter(v *visitFilters, cmp comparison, op Op) sql.Visitor {
	f, ok := cmp.Left().(*expression.GetField)
	if !ok {
		return nil
	}
	if _, ok := Indexed[f.Name()]; !ok {
		return nil
	}
	lit, ok := cmp.Right().(*expression.Literal)
	if !ok {
		return nil
	}
	val, _ := lit.Eval(nil, nil)
	fv, _, _ := f.Type().Convert(val)
	v.v.Fields = append(v.v.Fields, &FieldHint{
		Field: f.Name(),
		Op:    op,
		Value: fv,
	})
	return nil
}
