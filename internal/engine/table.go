package engine

import (
	"bytes"
	"io"
	"time"

	"github.com/apache/arrow/go/v14/parquet"
	"github.com/dgraph-io/badger/v4"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/oklog/ulid/v2"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/entry"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/must"
	v1 "github.com/vinceanalytics/vince/proto/v1"
	"google.golang.org/protobuf/proto"
)

type Table struct {
	Context
	name    string
	schema  sql.Schema
	columns []v1.Column
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
	prefix := []byte(keys.BlockIndex(t.name, "") + "/")
	o := badger.DefaultIteratorOptions
	o.Prefix = prefix
	txn := t.DB.NewTransaction(false)
	it := txn.Iter(db.IterOpts{
		Prefix:         prefix,
		PrefetchValues: true,
		PrefetchSize:   100,
	})
	it.Rewind()
	return &partitionIter{
		it:      it,
		txn:     txn,
		baseKey: prefix,
	}, nil
}

func (t *Table) PartitionRows(_ *sql.Context, p sql.Partition) (sql.RowIter, error) {
	x := p.(*Partition)
	var result []entry.ReadResult

	t.ReadBlock(x.Block, func(ras parquet.ReaderAtSeeker) {
		cols := t.columns
		if len(cols) == 0 {
			cols = Columns
		}
		result = entry.ReadColumns(
			entry.NewFileReader(ras),
			cols,
			x.RowGroups,
		)
	})
	return &rowIter{result: result}, nil
}

func (t *Table) WithProjections(colNames []string) sql.Table {
	m := make([]v1.Column, len(colNames))
	for i := range colNames {
		m[i] = v1.Column(v1.Column_value[colNames[i]])
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
	baseKey []byte
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

	var idx v1.Block_Index
	id := bytes.TrimPrefix(key, p.baseKey)

	must.One(p.it.Value(func(val []byte) error {
		return proto.Unmarshal(val, &idx)
	}))("failed decoding partition block index")

	// for now read all row groups
	var pat Partition
	pat.Block = ulid.MustParse(string(id))
	for i := range idx.RowGroupBitmap {
		pat.RowGroups = append(pat.RowGroups, i)
	}
	return &pat, nil
}

func (p *partitionIter) Close(*sql.Context) error {
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
	if p.pos < p.result[0].Len() {
		o := make(sql.Row, len(p.result))
		for i := range p.result {
			x := p.result[i].Col()
			if x <= v1.Column_timestamp {
				v := p.result[i].Value(p.pos).(int64)
				if x == v1.Column_timestamp {
					o[i] = time.UnixMilli(v)
				} else {
					o[i] = v
				}
				continue
			}
			o[i] = string(p.result[i].Value(p.pos).(parquet.ByteArray))
		}
		p.pos++
		return o, nil
	}
	return nil, io.EOF
}

func (p *rowIter) Close(*sql.Context) error { return nil }
