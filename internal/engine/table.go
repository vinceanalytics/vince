package engine

import (
	"fmt"
	"io"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/parquet-go/parquet-go"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
)

type Table struct {
	Context
	name        string
	schema      tableSchema
	projections []string
}

var _ sql.Table = (*Table)(nil)
var _ sql.ProjectedTable = (*Table)(nil)
var _ sql.IndexAddressable = (*Table)(nil)

func (t *Table) Name() string {
	return t.name
}

func (t *Table) String() string {
	return t.name
}

func (t *Table) Schema() sql.Schema {
	return t.schema.sql
}

func (t *Table) Collation() sql.CollationID {
	return sql.Collation_Default
}

func (t *Table) Partitions(*sql.Context) (sql.PartitionIter, error) {
	txn := t.DB.NewTransaction(false)
	it := txn.Iter(db.IterOpts{
		Prefix:         keys.BlockMetadata(t.name, ""),
		PrefetchValues: false,
	})
	it.Rewind()
	return &partitionIter{
		it:  it,
		txn: txn,
	}, nil
}

func (t *Table) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	x := partition.(*rangePartition)
	var record arrow.Record
	err := t.Reader.Read(ctx, x.Key(), func(f io.ReaderAt, size int64) error {
		r, err := parquet.OpenFile(f, size)
		if err != nil {
			return err
		}
		record, err = t.schema.read(ctx, r)
		return err
	})
	if err != nil {
		return nil, err
	}
	return newRecordIter(record), nil
}

func (t *Table) WithProjections(colNames []string) sql.Table {
	m := make([]storev1.Column, len(colNames))
	for i := range colNames {
		m[i] = storev1.Column(storev1.Column_value[colNames[i]])
	}
	return &Table{Context: t.Context,
		name:        t.name,
		schema:      createSchema(t.name, m),
		projections: colNames,
	}
}

func (t *Table) Projections() (o []string) {
	return t.projections
}

type Partition struct {
	Info blocksv1.BlockInfo
}

func (p *Partition) Key() []byte { return []byte(p.Info.Id) }

type partitionIter struct {
	it        db.Iter
	txn       db.Txn
	partition Partition
	started   bool
}

func (p *partitionIter) Next(*sql.Context) (sql.Partition, error) {
	if !p.started {
		p.started = true
		p.it.Rewind()
	} else {
		p.it.Next()
	}
	if !p.it.Valid() {
		return nil, io.EOF
	}
	err := p.it.Value(px.Decode(&p.partition.Info))
	if err != nil {
		return nil, fmt.Errorf("failed decoding block index err:%v", err)
	}
	return &p.partition, nil
}

func (p *partitionIter) Close(*sql.Context) error {
	p.it.Close()
	p.txn.Close()
	return nil
}

func (t *Table) getField(col string) (int, *sql.Column) {
	i := t.schema.sql.IndexOf(col, t.name)
	if i == -1 {
		return -1, nil
	}
	return i, t.schema.sql[i]
}

func (t *Table) GetIndexes(ctx *sql.Context) ([]sql.Index, error) {
	return []sql.Index{t.createIndex()}, nil
}

func (t *Table) createIndex() sql.Index {
	exprs := make([]sql.Expression, 0, len(t.schema.sql))
	exprsString := make([]string, 0, len(t.schema.sql))
	for _, column := range t.schema.sql {
		if _, ok := Indexed[column.Name]; !ok {
			continue
		}
		idx, field := t.getField(column.Name)
		ex := expression.NewGetFieldWithTable(idx, field.Type, t.name, field.Name, field.Nullable)
		exprs = append(exprs, ex)
		exprsString = append(exprsString, ex.String())
	}
	return &Index{
		DB:         "vince",
		Tbl:        t,
		TableName:  t.name,
		Exprs:      exprs,
		exprString: exprsString,
	}
}

type IndexedTable struct {
	*Table
	Lookup sql.IndexLookup
}

func (t *IndexedTable) LookupPartitions(ctx *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	filterExpr, filters, err := lookup.Index.(*Index).rangeFilterExpr(ctx, lookup.Ranges...)
	if err != nil {
		return nil, err
	}
	child, err := t.Table.Partitions(ctx)
	if err != nil {
		return nil, err
	}
	return rangePartitionIter{child: child.(*partitionIter),
		ranges:  filterExpr,
		filters: filters,
	}, nil
}

func (t *Table) IndexedAccess(i sql.IndexLookup) sql.IndexedTable {
	return &IndexedTable{Table: t, Lookup: i}
}

// PartitionRows implements the sql.PartitionRows interface.
func (t *IndexedTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	iter, err := t.Table.PartitionRows(ctx, partition)
	if err != nil {
		return nil, err
	}
	return iter, nil
}

// rangePartitionIter returns a partition that has range and table data access
type rangePartitionIter struct {
	child   *partitionIter
	ranges  sql.Expression
	filters *Filters
}

var _ sql.PartitionIter = (*rangePartitionIter)(nil)

func (i rangePartitionIter) Close(ctx *sql.Context) error {
	return i.child.Close(ctx)
}

func (i rangePartitionIter) Next(ctx *sql.Context) (sql.Partition, error) {
	part, err := i.child.Next(ctx)
	if err != nil {
		return nil, err
	}
	return &rangePartition{
		Partition: part.(*Partition),
		rang:      i.ranges,
	}, nil
}

type rangePartition struct {
	*Partition
	rang sql.Expression
}
