package engine

import (
	"errors"
	"fmt"
	"io"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/bits-and-blooms/bitset"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/parquet-go/parquet-go"
	"github.com/sirupsen/logrus"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	storev1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/b3"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
)

type Table struct {
	db          db.Provider
	reader      b3.Reader
	name        string
	schema      tableSchema
	projections []string
}

var _ sql.Table = (*Table)(nil)
var _ sql.ProjectedTable = (*Table)(nil)

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

func (t *Table) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	txn := t.db.NewTransaction(false)
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
	var record arrow.Record
	part := partition.(*Partition)
	if err := part.Valid(); err != nil {
		return nil, err
	}
	err := t.reader.Read(ctx, partition.Key(), func(f io.ReaderAt, size int64) error {
		r, err := parquet.OpenFile(f, size)
		if err != nil {
			return err
		}
		record, err = t.schema.read(ctx, r, part.RowGroups, part.Pages)
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
	return &Table{
		db:          t.db,
		reader:      t.reader,
		name:        t.name,
		schema:      createSchema(t.name, m),
		projections: colNames,
	}
}

func (t *Table) Projections() (o []string) {
	return t.projections
}

type Partition struct {
	Info      blocksv1.BlockInfo
	Index     []IndexFilter
	Values    []ValueFilter
	Expr      sql.Expression
	RowGroups []uint
	Pages     []*bitset.BitSet
	Range     bool
}

func (p *Partition) Valid() error {
	if !p.Range {
		return errors.New("non range queries are not supported for sites table")
	}
	var hasTs bool
	for i := range p.Index {
		if p.Index[i].Column() == v1.Column_timestamp {
			hasTs = true
			break
		}
	}
	if !hasTs {
		return errors.New("timestamp filter is required in the where clause")
	}
	return nil
}

func (p *Partition) Key() []byte { return []byte(p.Info.Id) }

type partitionIter struct {
	it        db.Iter
	txn       db.Txn
	partition Partition
	idx       blocksv1.ColumnIndex
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

func (p *partitionIter) readIndex(ctx *sql.Context, column v1.Column) *blocksv1.ColumnIndex {
	key := keys.BlockIndex(p.partition.Info.Domain, p.partition.Info.Id, column)
	err := p.txn.Get(key, px.Decode(&p.idx))
	if err != nil {
		ctx.GetLogger().
			WithFields(logrus.Fields{
				logrus.ErrorKey: err,
				"column":        column.String(),
				"block_id":      p.partition.Info.Id,
				"domain":        p.partition.Info.Domain,
			}).
			Warn("failed to retrieve column index")
		return nil
	}
	return &p.idx
}
