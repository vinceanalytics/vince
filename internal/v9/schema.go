package v9

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/polarsignals/frostdb"
	"github.com/polarsignals/frostdb/dynparquet"
	schemapb "github.com/polarsignals/frostdb/gen/proto/go/frostdb/schema/v1alpha1"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/spec"
)

type IngestFunc func(ts int64, uid, sid uint64, prop spec.Property, metric spec.Metric, key string, value int64)

type V9 struct {
	store *frostdb.ColumnStore
	db    *frostdb.DB
	table *frostdb.Table
	buf   *dynparquet.Buffer
	rows  []parquet.Row
}

func (v *V9) Ingest() (IngestFunc, func(context.Context)) {
	v.buf.Reset()
	v.rows = v.rows[:0]
	return v.Add, v.Save
}

func (v *V9) Add(ts int64, uid, sid uint64, prop spec.Property, metric spec.Metric, key string, value int64) {
	v.rows = append(v.rows, parquet.Row{
		parquet.Int64Value(ts),
		parquet.Int64Value(int64(uid)),
		parquet.Int64Value(int64(sid)),
		parquet.ByteArrayValue([]byte(prop.String())),
		parquet.ByteArrayValue([]byte(metric.String())),
		parquet.ByteArrayValue([]byte(key)),
		parquet.Int64Value(value),
	})
}

func (v *V9) Save(ctx context.Context) {
	_, err := v.table.InsertBuffer(ctx, v.buf)
	if err != nil {
		log.Get().Err(err).Msg("failed  saving buffer to table")
	}
}

func (v *V9) Close() error {
	return errors.Join(v.db.Close(), v.store.Close())
}

func Open(ctx context.Context) (*V9, error) {
	store, err := frostdb.New(
		frostdb.WithWAL(),
		frostdb.WithStoragePath(
			filepath.Join(config.Get(ctx).DataPath, "series"),
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

	table, err := db.Table("vince", frostdb.NewTableConfig(scheme))
	if err != nil {
		db.Close()
		store.Close()
		return nil, err
	}
	return &V9{store: store, db: db, table: table}, nil
}

var scheme = &schemapb.Schema{
	Name: "vince",
	Columns: []*schemapb.Column{
		{
			Name: "timestamp",
			StorageLayout: &schemapb.StorageLayout{
				Type: schemapb.StorageLayout_TYPE_INT64,
			},
		},
		{
			Name: "uid",
			StorageLayout: &schemapb.StorageLayout{
				Type: schemapb.StorageLayout_TYPE_INT64,
			},
		},
		{
			Name: "sid",
			StorageLayout: &schemapb.StorageLayout{
				Type: schemapb.StorageLayout_TYPE_INT64,
			},
		},
		{
			Name: "property",
			StorageLayout: &schemapb.StorageLayout{
				Type:     schemapb.StorageLayout_TYPE_STRING,
				Encoding: schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
			},
		},
		{
			Name: "metric",
			StorageLayout: &schemapb.StorageLayout{
				Type:     schemapb.StorageLayout_TYPE_STRING,
				Encoding: schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
			},
		},
		{
			Name: "key",
			StorageLayout: &schemapb.StorageLayout{
				Type:     schemapb.StorageLayout_TYPE_STRING,
				Encoding: schemapb.StorageLayout_ENCODING_RLE_DICTIONARY,
			},
		},
		{
			Name: "value",
			StorageLayout: &schemapb.StorageLayout{
				Type: schemapb.StorageLayout_TYPE_INT64,
			},
		},
	},
	SortingColumns: []*schemapb.SortingColumn{
		{
			Name:      "timestamp",
			Direction: schemapb.SortingColumn_DIRECTION_ASCENDING,
		},
		{
			Name:      "key",
			Direction: schemapb.SortingColumn_DIRECTION_ASCENDING,
		},
	},
}
