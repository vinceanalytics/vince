package db

import (
	"bytes"
	"context"

	"github.com/apache/arrow/go/v15/arrow"
	"github.com/apache/arrow/go/v15/arrow/array"
	"github.com/apache/arrow/go/v15/arrow/compute"
	"github.com/apache/arrow/go/v15/parquet"
	"github.com/apache/arrow/go/v15/parquet/file"
	"github.com/apache/arrow/go/v15/parquet/pqarrow"
	"github.com/vinceanalytics/vince/buffers"
	"github.com/vinceanalytics/vince/index"
)

type Part struct {
	GranuleID string
	*index.FileIndex
	Data arrow.Record
}

var _ index.Part = (*Part)(nil)

func NewRecord(ctx context.Context, db Storage, resource, id string, numRows int64, columns []int) (arrow.Record, error) {
	b := buffers.Bytes()
	defer b.Release()
	b.WriteString(resource)
	b.Write(slash)
	b.WriteString(id)
	b.Write(slash)
	b.Write(dataPath)
	var r arrow.Record
	err := db.Get(b.Bytes(), func(b []byte) error {
		f, err := file.NewParquetReader(bytes.NewReader(b),
			file.WithReadProps(parquet.NewReaderProperties(
				compute.GetAllocator(ctx),
			)),
		)
		if err != nil {
			return err
		}
		defer f.Close()
		pr, err := pqarrow.NewFileReader(f, pqarrow.ArrowReadProperties{
			BatchSize: numRows,
			Parallel:  true,
		},
			compute.GetAllocator(ctx),
		)
		if err != nil {
			return err
		}
		// There is only one row group
		table, err := pr.ReadRowGroups(ctx, columns, []int{0})
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		defer table.Release()
		r = tableToRecord(table)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewIndex(ctx context.Context, db Storage, resource, id string) (*index.FileIndex, error) {
	b := buffers.Bytes()
	defer b.Release()

	b.WriteString(resource)
	b.Write(slash)
	b.WriteString(id)
	var fdx *index.FileIndex
	err := db.Get(b.Bytes(), func(b []byte) error {
		var err error
		fdx, err = index.NewFileIndex(bytes.NewReader(b))
		return err
	})
	if err != nil {
		return nil, err
	}
	return fdx, err
}

func tableToRecord(table arrow.Table) arrow.Record {
	a := make([]arrow.Array, 0, table.NumCols())
	for i := 0; i < int(table.NumCols()); i++ {
		col := table.Column(i)
		// we read full batch so there is only one array in the chunk
		x := col.Data().Chunks()[0]
		x.Retain()
		a = append(a, x)
	}
	return array.NewRecord(
		table.Schema(), a, table.NumRows(),
	)
}
func (p *Part) ID() string {
	return p.GranuleID
}

func (p *Part) Record() arrow.Record {
	return p.Data
}

func (p *Part) Release() {
	p.Data.Release()
	p.Data = nil
}
