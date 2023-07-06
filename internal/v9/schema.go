package v9

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/polarsignals/frostdb"
	"github.com/polarsignals/frostdb/dynparquet"
	schemapb "github.com/polarsignals/frostdb/gen/proto/go/frostdb/schema/v1alpha1"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/spec"
)

type V9 struct {
	store         *frostdb.ColumnStore
	db            *frostdb.DB
	table         *frostdb.Table
	buf           *dynparquet.Buffer
	rows          []parquet.Row
	maxBufferSize int
}

type SaveAggregateFunc func(ctx context.Context, ts int64, uid, sid uint64, p spec.Property, key string, sum *entry.Aggregate) error

func (v *V9) Aggregate(ctx context.Context, ts int64, uid, sid uint64, p spec.Property, key string, sum *entry.Aggregate) error {
	return errors.Join(
		v.Add(ctx, ts, uid, sid, p, spec.Visitors, key, int64(sum.Visitors)),
		v.Add(ctx, ts, uid, sid, p, spec.Views, key, int64(sum.Views)),
		v.Add(ctx, ts, uid, sid, p, spec.Events, key, int64(sum.Events)),
		v.Add(ctx, ts, uid, sid, p, spec.Visits, key, int64(sum.Visits)),
		v.Add(ctx, ts, uid, sid, p, spec.BounceRates, key, int64(sum.BounceRates)),
		v.Add(ctx, ts, uid, sid, p, spec.VisitDurations, key, int64(sum.VisitDurations)),
	)
}

func (v *V9) Do(ctx context.Context, f func(SaveAggregateFunc) error) error {
	return errors.Join(
		f(v.Aggregate),
		v.save(ctx),
	)
}

func (v *V9) Add(ctx context.Context, ts int64, uid, sid uint64, prop spec.Property, metric spec.Metric, key string, value int64) error {
	if len(v.rows) > v.maxBufferSize {
		err := v.save(ctx)
		if err != nil {
			return err
		}
	}
	v.rows = append(v.rows, parquet.Row{
		parquet.Int64Value(ts),
		parquet.Int64Value(int64(uid)),
		parquet.Int64Value(int64(sid)),
		parquet.ByteArrayValue([]byte(prop.String())),
		parquet.ByteArrayValue([]byte(metric.String())),
		parquet.ByteArrayValue([]byte(key)),
		parquet.Int64Value(value),
	})
	return nil
}

func (v *V9) save(ctx context.Context) error {
	if len(v.rows) == 0 {
		return nil
	}
	_, err := v.buf.WriteRows(v.rows)
	if err != nil {
		return err
	}
	_, err = v.table.InsertBuffer(ctx, v.buf)
	if err != nil {
		return err
	}
	v.buf.Reset()
	v.rows = v.rows[:0]
	return nil
}

func (v *V9) Close() error {
	return errors.Join(v.save(context.Background()), v.db.Close(), v.store.Close())
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

	table, err := db.Table("vince", frostdb.NewTableConfig(scheme))
	if err != nil {
		db.Close()
		store.Close()
		return nil, err
	}
	buf, err := table.Schema().NewBuffer(nil)
	if err != nil {
		db.Close()
		store.Close()
		return nil, err
	}
	return &V9{store: store, db: db, table: table, buf: buf}, nil
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
			Name:      "uid",
			Direction: schemapb.SortingColumn_DIRECTION_ASCENDING,
		},
		{
			Name:      "sid",
			Direction: schemapb.SortingColumn_DIRECTION_ASCENDING,
		},
	},
}
