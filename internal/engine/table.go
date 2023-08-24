package engine

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"
)

type Table struct {
	name   string
	schema sql.Schema
}

var _ sql.Table = (*Table)(nil)

func (t *Table) Name() string {
	return t.name
}

func (t *Table) String() string {
	return t.name
}

func (t *Table) Schema() sql.Schema {
	return t.schema
}

func (t *Table) Collation() sql.CollationID {
	return sql.Collation_Default
}

func (t *Table) Partitions(*sql.Context) (sql.PartitionIter, error) {
	return &partitionIter{}, nil
}

func (t *Table) PartitionRows(*sql.Context, sql.Partition) (sql.RowIter, error) {
	return &rowIter{}, nil
}

type Partition struct {
	key []byte
}

func (p *Partition) Key() []byte { return p.key }

type partitionIter struct{}

func (p *partitionIter) Next(*sql.Context) (sql.Partition, error) {
	return nil, io.EOF
}

func (p *partitionIter) Close(*sql.Context) error { return nil }

type rowIter struct{}

func (p *rowIter) Next(*sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

func (p *rowIter) Close(*sql.Context) error { return nil }
