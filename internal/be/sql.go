package be

import (
	"fmt"

	"github.com/apache/arrow/go/v13/arrow"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/opcode"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/substrait-io/substrait-go/expr"
	"github.com/substrait-io/substrait-go/plan"
	"github.com/vinceanalytics/vince/internal/must"
)

type Parser struct {
	p *parser.Parser
}

func (p *Parser) Parse(sql string) (ParseResult, error) {
	asts, _, err := p.p.Parse(sql, "", "")
	if err != nil {
		return ParseResult{}, err
	}

	if len(asts) != 1 {
		return ParseResult{}, fmt.Errorf("cannot handle multiple asts, found %d", len(asts))
	}

	v := newASTVisitor()
	asts[0].Accept(v)
	if v.err != nil {
		return ParseResult{}, v.err
	}

	r, err := v.Build()
	if err != nil {
		return ParseResult{}, err
	}
	return ParseResult{Explain: v.explain, Plan: r}, nil
}

func NewParser() *Parser {
	return &Parser{p: parser.New()}
}

type ParseResult struct {
	Explain bool
	Plan    *plan.Plan
}

type astVisitor struct {
	explain bool
	plan    plan.Builder
	err     error
	rel     plan.Rel
	names   []string
	expr    []expr.Expression
	xb      expr.ExprBuilder
}

var _ ast.Visitor = (*astVisitor)(nil)

var _ ast.Visitor = (leaveFunc)(nil)

type leaveFunc func(n ast.Node) (node ast.Node, ok bool)

func (f leaveFunc) Enter(n ast.Node) (ast.Node, bool) {
	return n, false
}

func (f leaveFunc) Leave(n ast.Node) (ast.Node, bool) {
	return f(n)
}

func newASTVisitor() *astVisitor {
	return &astVisitor{
		plan: plan.NewBuilderDefault(),
		xb:   expr.ExprBuilder{Reg: extSet.GetSubstraitRegistry()},
	}
}

func (v *astVisitor) Build() (*plan.Plan, error) {
	return v.plan.Plan(v.rel, v.names)
}

func (v *astVisitor) Enter(n ast.Node) (nRes ast.Node, skipChildren bool) {
	switch e := n.(type) {
	case *ast.SelectStmt:
		var name *ast.TableName
		e.From.Accept(leaveFunc(func(n ast.Node) (node ast.Node, ok bool) {
			switch e := n.(type) {
			case *ast.TableName:
				name = e
				return nil, false
			}
			return n, true
		}))
		var columns []string
		e.Fields.Accept(leaveFunc(func(n ast.Node) (node ast.Node, ok bool) {
			switch e := n.(type) {
			case *ast.SelectField:
				if e.WildCard != nil {
					columns = all
					return nil, false
				}
			case *ast.ColumnName:
				columns = append(columns, e.Name.String())
				return n, true
			default:
			}
			return n, true
		}))
		root := Schema(columns...)
		table := v.plan.NamedScan([]string{name.Name.String()},
			root,
		)
		v.xb.BaseSchema = &root.Struct
		v.rel = table
		v.names = append(v.names[:0], columns...)
		if e.Where != nil {
			e.Where.Accept(leaveFunc(func(n ast.Node) (node ast.Node, ok bool) {
				switch e := n.(type) {
				case *ast.ColumnName:
					name := e.Name.String()
					for i := range v.names {
						if v.names[i] == name {
							v.expr = append(v.expr,
								must.Must(
									v.plan.RootFieldRef(table, int32(i)),
								)(),
							)
							return n, true
						}
					}
					v.err = fmt.Errorf("column %q does not exist", name)
					return nil, false
				// case *test_driver.ValueExpr:
				// 	var lit expr.Literal
				// 	switch x := e.GetValue().(type) {
				// 	case int64:
				// 		lit = expr.NewPrimitiveLiteral(x, false)
				// 	case float32:
				// 		lit = expr.NewPrimitiveLiteral(x, false)
				// 	case float64:
				// 		lit = expr.NewPrimitiveLiteral(x, false)
				// 	case string:
				// 		lit = expr.NewPrimitiveLiteral(x, false)
				// 	default:
				// 		v.err = fmt.Errorf("data type %v is not supported", x)
				// 		return nil, false
				// 	}
				// 	v.expr = append(v.expr, lit)

				case *ast.BinaryOperationExpr:
					right, n := pop(v.expr)
					left, n := pop(n)
					v.expr = n
					var op string
					switch e.Op {
					case opcode.GT:
						op = "greater"
					case opcode.LT:
						op = "less"
					case opcode.GE:
						op = "greater_equal"
					case opcode.LE:
						op = "less_equal"
					case opcode.EQ:
						op = "equal"
					case opcode.NE:
						op = "not_equal"
						v.err = fmt.Errorf("operator %q is not supported", e.Op)
						return nil, false
					}
					f, err := v.callScalar(op, left, right)
					if err != nil {
						v.err = err
						return nil, false
					}
					v.expr = append(v.expr, f)
				default:
					fmt.Printf("WHERE %#T\n", e)
				}
				return n, true
			}))
			if v.err != nil {
				return nil, true
			}
			cond, n := pop(v.expr)
			v.expr = n
			f, err := v.plan.Filter(table, cond)
			if err != nil {
				v.err = err
				return nil, true
			}
			v.rel = f
		}

	// 	// The SelectStmt is handled in during pre-visit given that it has many
	// 	// clauses we need to handle independently (e.g. a group by with a
	// 	// filter).
	// 	if expr.Where != nil {
	// 		expr.Where.Accept(v)
	// 		lastExpr, newExprs := pop(v.exprStack)
	// 		v.exprStack = newExprs
	// 		v.builder = v.builder.Filter(lastExpr)
	// 	}
	// 	expr.Fields.Accept(v)
	// 	switch {
	// 	case expr.GroupBy != nil:
	// 		expr.GroupBy.Accept(v)
	// 		var agg []logicalplan.Expr
	// 		var groups []logicalplan.Expr

	// 		for _, expr := range v.exprStack {
	// 			switch expr.(type) {
	// 			case *logicalplan.AliasExpr, *logicalplan.AggregationFunction:
	// 				agg = append(agg, expr)
	// 			default:
	// 				groups = append(groups, expr)
	// 			}
	// 		}
	// 		v.builder = v.builder.Aggregate(agg, groups)
	// 	case expr.Distinct:
	// 		v.builder = v.builder.Distinct(v.exprStack...)
	// 	default:
	// 		v.builder = v.builder.Project(v.exprStack...)
	// 	}
	// 	return n, true
	default:
		fmt.Printf("=> %#T\n", e)
	}
	return n, false
}

func (v *astVisitor) Leave(n ast.Node) (nRes ast.Node, ok bool) {
	if err := v.leaveImpl(n); err != nil {
		v.err = err
		return n, false
	}
	return n, true
}

func (v *astVisitor) leaveImpl(n ast.Node) error {
	switch expr := n.(type) {
	case *ast.SelectStmt:
		// Handled in Enter.
		return nil
	// case *ast.ExplainStmt:
	// 	v.explain = true
	// 	return nil
	// case *ast.AggregateFuncExpr:
	// 	// At this point, the child node is the column name, so it has just been
	// 	// added to exprs.
	// 	lastExpr := len(v.exprStack) - 1
	// 	switch strings.ToLower(expr.F) {
	// 	case "count":
	// 		v.exprStack[lastExpr] = logicalplan.Count(v.exprStack[lastExpr])
	// 	case "sum":
	// 		v.exprStack[lastExpr] = logicalplan.Sum(v.exprStack[lastExpr])
	// 	case "min":
	// 		v.exprStack[lastExpr] = logicalplan.Min(v.exprStack[lastExpr])
	// 	case "max":
	// 		v.exprStack[lastExpr] = logicalplan.Max(v.exprStack[lastExpr])
	// 	case "avg":
	// 		v.exprStack[lastExpr] = logicalplan.Avg(v.exprStack[lastExpr])
	// 	default:
	// 		return fmt.Errorf("unhandled aggregate function %s", expr.F)
	// 	}
	// case *ast.BinaryOperationExpr:
	// 	// Note that we're resolving exprs as a stack, so the last two
	// 	// expressions are the leaf expressions.
	// 	rightExpr, newExprs := pop(v.exprStack)
	// 	leftExpr, newExprs := pop(newExprs)
	// 	v.exprStack = newExprs

	// 	var frostDBOp logicalplan.Op
	// 	switch expr.Op {
	// 	case opcode.GT:
	// 		frostDBOp = logicalplan.OpGt
	// 	case opcode.LT:
	// 		frostDBOp = logicalplan.OpLt
	// 	case opcode.GE:
	// 		frostDBOp = logicalplan.OpGtEq
	// 	case opcode.LE:
	// 		frostDBOp = logicalplan.OpLtEq
	// 	case opcode.EQ:
	// 		frostDBOp = logicalplan.OpEq
	// 	case opcode.NE:
	// 		frostDBOp = logicalplan.OpNotEq
	// 	case opcode.LogicAnd:
	// 		v.exprStack = append(v.exprStack, logicalplan.And(leftExpr, rightExpr))
	// 		return nil
	// 	case opcode.LogicOr:
	// 		v.exprStack = append(v.exprStack, logicalplan.Or(leftExpr, rightExpr))
	// 		return nil
	// 	}
	// 	v.exprStack = append(v.exprStack, &logicalplan.BinaryExpr{
	// 		Left:  logicalplan.Col(leftExpr.Name()),
	// 		Op:    frostDBOp,
	// 		Right: rightExpr,
	// 	})
	// case *ast.ColumnName:
	// 	colName := columnNameToString(expr)
	// 	var col logicalplan.Expr
	// 	if _, ok := v.dynColNames[colName]; ok {
	// 		col = logicalplan.DynCol(colName)
	// 	} else {
	// 		col = logicalplan.Col(colName)
	// 	}
	// 	v.exprStack = append(v.exprStack, col)
	// case *test_driver.ValueExpr:
	// 	switch logicalplan.Literal(expr.GetValue()).Name() { // NOTE: special case for boolean fields since the mysql parser doesn't support booleans as a type
	// 	case "true":
	// 		v.exprStack = append(v.exprStack, logicalplan.Literal(true))
	// 	case "false":
	// 		v.exprStack = append(v.exprStack, logicalplan.Literal(false))
	// 	default:
	// 		v.exprStack = append(v.exprStack, logicalplan.Literal(expr.GetValue()))
	// 	}
	// case *ast.SelectField:
	// 	if as := expr.AsName.String(); as != "" {
	// 		lastExpr := len(v.exprStack) - 1
	// 		v.exprStack[lastExpr] = v.exprStack[lastExpr].(*logicalplan.AggregationFunction).Alias(as) // TODO should probably just be an alias expr and not from an aggregate function
	// 	}
	// case *ast.PatternRegexpExpr:
	// 	rightExpr, newExprs := pop(v.exprStack)
	// 	leftExpr, newExprs := pop(newExprs)
	// 	v.exprStack = newExprs

	// 	e := &logicalplan.BinaryExpr{
	// 		Left:  logicalplan.Col(leftExpr.Name()),
	// 		Op:    logicalplan.OpRegexMatch,
	// 		Right: rightExpr,
	// 	}
	// 	if expr.Not {
	// 		e.Op = logicalplan.OpRegexNotMatch
	// 	}
	// 	v.exprStack = append(v.exprStack, e)
	// case *ast.FieldList, *ast.ColumnNameExpr, *ast.GroupByClause, *ast.ByItem, *ast.RowExpr,
	// 	*ast.ParenthesesExpr:
	// 	// Deliberate pass-through nodes.
	// case *ast.FuncCallExpr:
	// 	switch expr.FnName.String() {
	// 	case ast.Second:
	// 		// This is pretty hacky and only fine because it's in the test only.
	// 		left, right := pop(v.exprStack)
	// 		var exprStack []logicalplan.Expr
	// 		exprStack = append(exprStack, right...)
	// 		switch l := left.(type) {
	// 		case *logicalplan.LiteralExpr:
	// 			val := l.Value.(*scalar.Int64)
	// 			duration := time.Duration(val.Value) * time.Second
	// 			exprStack = append(exprStack, logicalplan.Duration(duration))
	// 			v.exprStack = exprStack
	// 		}
	// 	default:
	// 		return fmt.Errorf("unhandled func call: %s", expr.FnName.String())
	// 	}
	default:
		fmt.Printf("==> %T\n", expr)
		return nil
	}
}

func columnNameToString(c *ast.ColumnName) string {
	// Note that in SQL labels.label2 is interpreted as referencing
	// the label2 column of a table called labels. In our case,
	// these are dynamic columns, which is why the table name is
	// accessed here.
	colName := ""
	if c.Table.String() != "" {
		colName = c.Table.String() + "."
	}
	colName += c.Name.String()
	return colName
}

func pop[T any](s []T) (T, []T) {
	lastIdx := len(s) - 1
	return s[lastIdx], s[:lastIdx]
}

func (v *astVisitor) callScalar(name string, args ...expr.Expression) (*expr.ScalarFunction, error) {
	conv, ok := extSet.GetArrowRegistry().GetArrowToSubstrait(name)
	if !ok {
		return nil, arrow.ErrNotFound
	}
	id, opts, err := conv(name)
	if err != nil {
		return nil, err
	}

	a := make([]expr.FuncArgBuilder, len(args))
	for i := range args {
		switch e := args[i].(type) {
		case expr.Literal:
			a[i] = v.xb.Literal(e)
		case *expr.FieldReference:
			a[i] = v.xb.RootRef(e.Reference)
		default:
			return nil, fmt.Errorf("unsupported argument type %#T", e)
		}
	}
	return v.xb.ScalarFunc(id, opts...).Args(a...).Build()
}
