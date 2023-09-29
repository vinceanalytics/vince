package engine

import (
	"container/heap"
	"errors"
	"fmt"
	"io"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/bits-and-blooms/bitset"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/parquet-go/parquet-go"
	"github.com/sirupsen/logrus"
	blocksv1 "github.com/vinceanalytics/vince/gen/proto/go/vince/blocks/v1"
	v1 "github.com/vinceanalytics/vince/gen/proto/go/vince/store/v1"
	"github.com/vinceanalytics/vince/internal/db"
	"github.com/vinceanalytics/vince/internal/engine/session"
	"github.com/vinceanalytics/vince/internal/keys"
	"github.com/vinceanalytics/vince/internal/px"
)

type SitesTable struct {
	schema      tableSchema
	projections []string
}

var _ sql.Table = (*SitesTable)(nil)
var _ sql.ProjectedTable = (*SitesTable)(nil)

func (*SitesTable) Name() string                 { return SitesTableName }
func (*SitesTable) String() string               { return SitesTableName }
func (t *SitesTable) Schema() sql.Schema         { return t.schema.sql }
func (t *SitesTable) Collation() sql.CollationID { return sql.Collation_Default }

func (t *SitesTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	db := session.Get(ctx).DB()
	return &partitionIter{
		txn: db.NewTransaction(false),
	}, nil
}

func (t *SitesTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	var record arrow.Record
	part := partition.(*Partition)
	if err := part.Valid(); err != nil {
		return nil, err
	}
	reader := session.Get(ctx).B3()
	err := reader.Read(ctx, partition.Key(), func(f io.ReaderAt, size int64) error {
		r, err := parquet.OpenFile(f, size)
		if err != nil {
			return err
		}
		record, err = t.schema.read(ctx, part.Info.Domain, r, part.RowGroups, part.Pages)
		if err != nil {
			return err
		}
		record, err = applyValueFilter(ctx, part.Filters.Values, record)
		return err
	})
	if err != nil {
		return nil, err
	}
	return newRecordIter(record), nil
}

func (t *SitesTable) WithProjections(colNames []string) sql.Table {
	return &SitesTable{
		schema:      createSchema(colNames),
		projections: colNames,
	}
}

func (t *SitesTable) Projections() (o []string) {
	return t.projections
}

type Partition struct {
	Info      blocksv1.BlockInfo
	Filters   FilterContext
	RowGroups []uint
	Pages     []*bitset.BitSet
}

func (p *Partition) Valid() error {
	var hasTs bool
	for i := range p.Filters.Index {
		if p.Filters.Index[i].Column() == v1.Column_timestamp {
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

func (p *partitionIter) Next(ctx *sql.Context) (sql.Partition, error) {
	if !p.started {
		if p.partition.Filters.Domains.Len() > 0 {
			site := heap.Pop(&p.partition.Filters.Domains)
			p.it = p.txn.Iter(db.IterOpts{
				Prefix: keys.BlockMetadata(site.(string), ""),
			})
		} else {
			// all sites
			p.it = p.txn.Iter(db.IterOpts{
				Prefix: keys.BlockMetadataPrefix(),
			})
		}
		p.started = true
		p.it.Rewind()
	} else {
		p.it.Next()
	}
	if !p.it.Valid() {
		if p.partition.Filters.Domains.Len() > 0 {
			// we still have domains to work with
			p.it.Close()
			p.started = false
			return p.Next(ctx)
		}
		return nil, io.EOF
	}
	err := p.it.Value(px.Decode(&p.partition.Info))
	if err != nil {
		return nil, fmt.Errorf("failed decoding block index err:%v", err)
	}
	rs, err := buildIndexFilter(ctx, p.partition.Filters.Index, p.readIndex)
	if err != nil {
		if errors.Is(err, ErrSkipBlock) {
			return p.Next(ctx)
		}
		if !errors.Is(err, ErrNoFilter) {
			ctx.GetLogger().WithError(err).Warn("failed to build index filter")
		}
	} else {
		p.partition.RowGroups = rs.RowGroups
		p.partition.Pages = rs.Pages
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
