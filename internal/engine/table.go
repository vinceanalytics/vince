package engine

import (
	"bytes"
	"io"

	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
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

func (t *Table) PartitionRows(ctx *sql.Context, p sql.Partition) (sql.RowIter, error) {
	x := p.(*Partition)
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
