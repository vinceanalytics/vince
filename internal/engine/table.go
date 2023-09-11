package engine

import (
	"bytes"
	"io"

	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/oklog/ulid/v2"
	"github.com/parquet-go/parquet-go"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/mem"
	"github.com/vinceanalytics/vince/internal/must"
	"google.golang.org/protobuf/proto"
)

type Table struct {
	Context
	name    string
	schema  sql.Schema
	columns []storev1.Column
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
	if len(t.columns) > 0 {
		return Schema(t.name, t.columns)
	}
	return t.schema
}

func (t *Table) Collation() sql.CollationID {
	return sql.Collation_Default
}

func (t *Table) Partitions(*sql.Context) (sql.PartitionIter, error) {
	key := keys.BlockIndex(t.name, "")
	o := badger.DefaultIteratorOptions
	o.Prefix = key.Bytes()
	txn := t.DB.NewTransaction(false)
	it := txn.Iter(db.IterOpts{
		Prefix:         key.Bytes(),
		PrefetchValues: true,
		PrefetchSize:   100,
	})
	it.Rewind()
	return &partitionIter{
		it:      it,
		txn:     txn,
		baseKey: key,
	}, nil
}

func (t *Table) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	var p *Partition
	switch e := partition.(type) {
	case *rangePartition:
		p = e.Partition
	case *Partition:
		p = e
	}
	x := p
	var result []entry.ReadResult
	t.Reader.Read(ctx, x.Block, func(f io.ReaderAt, size int64) error {
		r, err := parquet.OpenFile(f, size)
		if err != nil {
			return err
		}
		cols := t.columns
		if len(cols) == 0 {
			cols = Columns
		}
		result = append(result, mem.ReadColumns(r, cols, x.RowGroups)...)
		return nil
	})
	return &rowIter{result: result}, nil
}

func (t *Table) WithProjections(colNames []string) sql.Table {
	m := make([]storev1.Column, len(colNames))
	for i := range colNames {
		m[i] = storev1.Column(storev1.Column_value[colNames[i]])
	}
	return &Table{Context: t.Context,
		name:    t.name,
		columns: m,
		schema:  Schema(t.name, m),
	}
}

func (t *Table) Projections() (o []string) {
	o = make([]string, len(t.columns))
	for i := range t.columns {
		o[i] = t.columns[i].String()
	}
	return
}

type Partition struct {
	RowGroups []int
	Block     ulid.ULID
}

func (p *Partition) Key() []byte { return p.Block[:] }

type partitionIter struct {
	it      db.Iter
	txn     db.Txn
	baseKey *keys.Key
	started bool
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
	key := p.it.Key()

	var idx blocksv1.BlockIndex
	id := bytes.TrimPrefix(key, p.baseKey.Bytes())

	must.One(p.it.Value(func(val []byte) error {
		return proto.Unmarshal(val, &idx)
	}))("failed decoding partition block index")

	// for now read all row groups
	var pat Partition
	pat.Block = ulid.MustParse(string(id))
	for i := range idx.Bloom {
		pat.RowGroups = append(pat.RowGroups, i)
	}
	return &pat, nil
}

func (p *partitionIter) Close(*sql.Context) error {
	p.baseKey.Release()
	p.baseKey = nil
	p.it.Close()
	p.txn.Close()
	return nil
}

type rowIter struct {
	result []entry.ReadResult
	pos    int
}

func (p *rowIter) Next(*sql.Context) (sql.Row, error) {
	if len(p.result) == 0 {
		return nil, io.EOF
	}
	rows := p.result[0].Len()
	if p.pos < rows {
		o := make(sql.Row, len(p.result))
		for i := range p.result {
			o[i] = p.result[i].Value(p.pos)
		}
		p.pos++
		return o, nil
	}
	return nil, io.EOF
}

func (p *rowIter) Close(*sql.Context) error { return nil }

func (t *Table) getField(col string) (int, *sql.Column) {
	i := t.schema.IndexOf(col, t.name)
	if i == -1 {
		return -1, nil
	}
	return i, t.schema[i]
}

func (t *Table) GetIndexes(ctx *sql.Context) ([]sql.Index, error) {
	return []sql.Index{t.createIndex()}, nil
}

func (t *Table) createIndex() sql.Index {
	exprs := make([]sql.Expression, len(t.schema))
	for i, column := range t.schema {
		if !Indexed[column.Name] {
			continue
		}
		idx, field := t.getField(column.Name)
		exprs[i] = expression.NewGetFieldWithTable(idx, field.Type, t.name, field.Name, field.Nullable)
	}
	return &Index{
		DB:        "vince",
		Tbl:       t,
		TableName: t.name,
		Exprs:     exprs,
	}
}

type IndexedTable struct {
	*Table
	Lookup sql.IndexLookup
}

func (t *IndexedTable) LookupPartitions(ctx *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	filter, err := lookup.Index.(*Index).rangeFilterExpr(ctx, lookup.Ranges...)
	if err != nil {
		return nil, err
	}
	child, err := t.Table.Partitions(ctx)
	if err != nil {
		return nil, err
	}
	return rangePartitionIter{child: child.(*partitionIter), ranges: filter}, nil
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
	child  *partitionIter
	ranges sql.Expression
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
