package v9

import (
	"context"
	"errors"
	"path/filepath"
	"sync"

	"github.com/polarsignals/frostdb"
	"github.com/polarsignals/frostdb/dynparquet"
	schemapb "github.com/polarsignals/frostdb/gen/proto/go/frostdb/schema/v1alpha2"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/spec"
)

type V9 struct {
	store *frostdb.ColumnStore
	db    *frostdb.DB
}

type rowsBuffer struct {
	buf           *dynparquet.Buffer
	rows          []parquet.Row
	maxBufferSize int
}

func (r *rowsBuffer) reset() {
	r.buf.Reset()
	r.rows = r.rows[:0]
}

func (r *rowsBuffer) release() {
	r.reset()
	rowsBufferPool.Put(r)
}

func newRowsBuffer() *rowsBuffer {
	return rowsBufferPool.Get().(*rowsBuffer)
}

var rowsBufferPool = &sync.Pool{
	New: func() any {
		return &rowsBuffer{
			maxBufferSize: 1 << 10,
			buf:           must(schema.NewBufferV2()),
			rows:          make([]parquet.Row, 0, 1<<10),
		}
	},
}

type SaveAggregateFunc func(ctx context.Context, ts int64, p spec.Property, key string, sum *entry.Aggregate) error

func (v *V9) agg(table *frostdb.Table, r *rowsBuffer) SaveAggregateFunc {
	return func(ctx context.Context, ts int64, p spec.Property, key string, sum *entry.Aggregate) error {
		return errors.Join(
			v.Add(ctx, table, r, ts, p, spec.Visitors, key, int64(sum.Visitors)),
			v.Add(ctx, table, r, ts, p, spec.Views, key, int64(sum.Views)),
			v.Add(ctx, table, r, ts, p, spec.Events, key, int64(sum.Events)),
			v.Add(ctx, table, r, ts, p, spec.Visits, key, int64(sum.Visits)),
			v.Add(ctx, table, r, ts, p, spec.BounceRates, key, int64(sum.BounceRates)),
			v.Add(ctx, table, r, ts, p, spec.VisitDurations, key, int64(sum.VisitDurations)),
		)
	}
}

func (v *V9) Do(ctx context.Context, domain string, f func(SaveAggregateFunc) error) error {
	table, err := v.db.GetTable(domain)
	if err != nil {
		// Only a single error is returned. We are sure that err is for the missing
		// table.
		table, err = v.db.Table(domain, frostdb.NewTableConfig(scheme))
		if err != nil {
			return err
		}
	}
	r := newRowsBuffer()
	defer r.release()
	return errors.Join(
		f(v.agg(table, r)),
		v.save(ctx, table, r),
	)
}

func (v *V9) Add(ctx context.Context, table *frostdb.Table, r *rowsBuffer, ts int64, prop spec.Property, metric spec.Metric, key string, value int64) error {
	if len(r.rows) >= r.maxBufferSize {
		err := v.save(ctx, table, r)
		if err != nil {
			return err
		}
	}
	r.rows = append(r.rows, parquet.Row{
		parquet.Int64Value(ts),
		parquet.ByteArrayValue([]byte(prop.String())),
		parquet.ByteArrayValue([]byte(metric.String())),
		parquet.ByteArrayValue([]byte(key)),
		parquet.Int64Value(value),
	})
	return nil
}

func (v *V9) save(ctx context.Context, table *frostdb.Table, r *rowsBuffer) error {
	if len(r.rows) == 0 {
		return nil
	}
	_, err := r.buf.WriteRows(r.rows)
	if err != nil {
		return err
	}
	_, err = table.InsertBuffer(ctx, r.buf)
	if err != nil {
		return err
	}
	r.reset()
	return nil
}

func (v *V9) Close() error {
	return errors.Join(v.db.Close(), v.store.Close())
}

func Open(ctx context.Context, dataPath string) (*V9, error) {
	store, err := frostdb.New(
		frostdb.WithWAL(),
		frostdb.WithStoragePath(
			filepath.Join(dataPath, "series"),
		),
	)
	if err != nil {
		return nil, err
	}
	db, err := store.DB(ctx, "vince")
	if err != nil {
		store.Close()
		return nil, err
	}

	return &V9{store: store, db: db}, nil
}

var schema = must(dynparquet.SchemaFromDefinition(scheme))

func must[T any](v T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return v
}

var scheme = &schemapb.Schema{
	Root: &schemapb.Group{
		Name: "site_stats",
		Nodes: []*schemapb.Node{
			{
				Type: &schemapb.Node_Leaf{
					Leaf: &schemapb.Leaf{Name: "timestamp", StorageLayout: &schemapb.StorageLayout{
						Type:        schemapb.StorageLayout_TYPE_INT64,
						Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
					}},
				},
			},
			{
				Type: &schemapb.Node_Leaf{
					Leaf: &schemapb.Leaf{Name: "segment", StorageLayout: &schemapb.StorageLayout{
						Type:        schemapb.StorageLayout_TYPE_STRING,
						Encoding:    schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
						Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
					}},
				},
			},
			{
				Type: &schemapb.Node_Leaf{
					Leaf: &schemapb.Leaf{Name: "metric", StorageLayout: &schemapb.StorageLayout{
						Type:        schemapb.StorageLayout_TYPE_STRING,
						Encoding:    schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
						Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
					}},
				},
			},
			{
				Type: &schemapb.Node_Leaf{
					Leaf: &schemapb.Leaf{Name: "key", StorageLayout: &schemapb.StorageLayout{
						Type:        schemapb.StorageLayout_TYPE_STRING,
						Encoding:    schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
						Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
					}},
				},
			},
			{
				Type: &schemapb.Node_Leaf{
					Leaf: &schemapb.Leaf{Name: "value", StorageLayout: &schemapb.StorageLayout{
						Type:        schemapb.StorageLayout_TYPE_INT64,
						Compression: schemapb.StorageLayout_COMPRESSION_ZSTD,
					}},
				},
			},
		},
	},
	SortingColumns: []*schemapb.SortingColumn{
		{
			Path:      "timestamp",
			Direction: schemapb.SortingColumn_DIRECTION_ASCENDING,
		},
	},
}
