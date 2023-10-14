package engine

import (
	"container/heap"
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	v1 "github.com/vinceanalytics/proto/gen/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/engine/session"
)

type Index struct {
	TableName  string
	Exprs      []sql.Expression
	exprString []string
	PrefixLens []uint16
}

var _ sql.Index = (*Index)(nil)
var _ sql.IndexAddressable = (*eventsTable)(nil)

func (idx *Index) Database() string                    { return "vince" }
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
func (idx *Index) ID() string                          { return "vince_fields_idx" }

func (idx *Index) PrefixLengths() []uint16 { return idx.PrefixLens }

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

type FilterContext struct {
	Domains Domains
	Index   []IndexFilter
	Values  []ValueFilter
}

func (i *FilterContext) build(col v1.Column, lo, hi any, op Op) {
	if f := buildIndexFilterMatch(col, lo, hi, op); f != nil {
		i.Index = append(i.Index, f)
	}
	if col == v1.Column_domain && op == Eq {
		i.Domains = append(i.Domains, lo.(string))
	}
	i.Values = append(i.Values, buildValueFilterMatch(col, lo, hi, op))
}

func (idx *Index) build(ctx *sql.Context,
	ranges ...sql.Range) (o FilterContext, err error) {
	if len(ranges) == 0 {
		return
	}
	exprs := idx.Exprs
	for rangIdx, rang := range ranges {
		if len(exprs) < len(rang) {
			err = fmt.Errorf("expected different key count: exprs(%d) < (ranges[%d])(%d)", len(exprs), rangIdx, len(rang))
			return
		}
		for i, rce := range rang {
			field := indexedField(exprs[i])
			lo, hi := bounds(rce)
			switch rce.Type() {
			case sql.RangeType_EqualNull:
				// All indexed fields are non nullable
			case sql.RangeType_GreaterThan:
				if sql.RangeCutIsBinding(rce.LowerBound) {
					o.build(field, lo, hi, Gt)
				}
			case sql.RangeType_GreaterOrEqual:
				o.build(field, lo, hi, GtEg)
			case sql.RangeType_LessThanOrNull:
				o.build(field, lo, hi, Lt)
			case sql.RangeType_LessOrEqualOrNull:
				o.build(field, lo, hi, LtEq)
			case sql.RangeType_ClosedClosed:
				o.build(field, lo, hi, Eq)
			case sql.RangeType_OpenOpen:
				if sql.RangeCutIsBinding(rce.LowerBound) {
					o.build(field, lo, hi, Gt)
				} else {
					// Lower bound is (NULL, ...)
					o.build(field, lo, hi, Lt)
				}
			case sql.RangeType_OpenClosed:
				if sql.RangeCutIsBinding(rce.LowerBound) {
					o.build(field, lo, hi, Lt)
					o.build(field, lo, hi, GtEg)
				} else {
					o.build(field, lo, hi, LtEq)
				}
			case sql.RangeType_ClosedOpen:
				o.build(field, lo, hi, GtEg)
				o.build(field, lo, hi, Lt)
			}
		}
	}
	return
}

func bounds(rce sql.RangeColumnExpr) (lo, hi any) {
	switch e := rce.LowerBound.(type) {
	case sql.Below:
		lo = e.Key
	case sql.Above:
		lo = e.Key
	}
	switch e := rce.UpperBound.(type) {
	case sql.Below:
		hi = e.Key
	case sql.Above:
		hi = e.Key
	}
	return
}

func indexedField(expr sql.Expression) v1.Column {
	f := expr.(*expression.GetField)
	return Indexed[f.Name()]
}

func (t *eventsTable) GetIndexes(ctx *sql.Context) ([]sql.Index, error) {
	return []sql.Index{t.createIndex()}, nil
}

func (t *eventsTable) createIndex() sql.Index {
	exprs := make([]sql.Expression, 0, len(t.schema.sql))
	exprsString := make([]string, 0, len(t.schema.sql))
	for _, column := range t.schema.sql {
		if _, ok := Indexed[column.Name]; !ok {
			continue
		}
		idx, field := t.getField(column.Name)
		ex := expression.NewGetFieldWithTable(idx, field.Type, eventsTableName, field.Name, field.Nullable)
		exprs = append(exprs, ex)
		exprsString = append(exprsString, ex.String())
	}
	return &Index{
		TableName:  eventsTableName,
		Exprs:      exprs,
		exprString: exprsString,
	}
}

func (t *eventsTable) getField(col string) (int, *sql.Column) {
	i := t.schema.sql.IndexOf(col, eventsTableName)
	if i == -1 {
		return -1, nil
	}
	return i, t.schema.sql[i]
}

type IndexedTable struct {
	*eventsTable
	Lookup sql.IndexLookup
}

func (t *IndexedTable) LookupPartitions(ctx *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	o, err := lookup.Index.(*Index).build(ctx, lookup.Ranges...)
	if err != nil {
		return nil, err
	}
	heap.Init(&o.Domains)
	db := session.Get(ctx).DB()
	return &partitionIter{
		txn: db.NewTransaction(false),
		partition: eventPartition{
			Filters: o,
		},
	}, nil

}

func (t *eventsTable) IndexedAccess(i sql.IndexLookup) sql.IndexedTable {
	return &IndexedTable{eventsTable: t, Lookup: i}
}

// PartitionRows implements the sql.PartitionRows interface.
func (t *IndexedTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	iter, err := t.eventsTable.PartitionRows(ctx, partition)
	if err != nil {
		return nil, err
	}
	return iter, nil
}

type Domains []string

func (h Domains) Len() int           { return len(h) }
func (h Domains) Less(i, j int) bool { return h[i] < h[j] }
func (h Domains) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *Domains) Push(x any) {
	*h = append(*h, x.(string))
}

func (h *Domains) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
