package timeseries

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"time"

	"github.com/apache/arrow/go/v13/arrow/memory"
	"github.com/polarsignals/frostdb"
	"github.com/polarsignals/frostdb/dynparquet"
	"github.com/polarsignals/frostdb/query"
	lo "github.com/polarsignals/frostdb/query/logicalplan"
	"github.com/segmentio/parquet-go"
	"github.com/thanos-io/objstore/providers/filesystem"
	"github.com/vinceanalytics/vince/pkg/entry"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/spec"
)

func Save(ctx context.Context, b *Buffer) {
	err := store(ctx).Save(ctx, b.domain, b.rows)
	if err != nil {
		log.Get().Err(err).Msg("failed saving events to storage")
	}
	b.Release()
}

type Store struct {
	store *frostdb.ColumnStore
	db    *frostdb.DB
}

type rowsBuffer struct {
	buf *dynparquet.Buffer
}

func (r *rowsBuffer) reset() {
	r.buf.Reset()
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
			buf: entry.SchemaBuffer(),
		}
	},
}

func (v *Store) Save(ctx context.Context, domain string, rows []parquet.Row) error {
	table, err := v.db.GetTable(domain)
	if err != nil {
		// Only a single error is returned. We are sure that err is for the missing
		// table.
		table, err = v.db.Table(domain, frostdb.NewTableConfig(entry.Scheme))
		if err != nil {
			return err
		}
	}
	r := newRowsBuffer()
	defer r.release()
	_, err = r.buf.WriteRows(rows)
	if err != nil {
		return err
	}
	_, err = table.InsertBuffer(ctx, r.buf)
	return err
}

func (v *Store) Close() error {
	return errors.Join(v.db.Close(), v.store.Close())
}

func OpenStore(ctx context.Context, dataPath string) (*Store, error) {
	bucket, err := filesystem.NewBucket(filepath.Join(dataPath, "buckets"))
	if err != nil {
		return nil, err
	}
	o := frostdb.NewDefaultObjstoreBucket(bucket)
	store, err := frostdb.New(
		frostdb.WithWAL(),
		frostdb.WithStoragePath(filepath.Join(dataPath, "store")),
		frostdb.WithReadWriteStorage(o),
	)
	if err != nil {
		return nil, err
	}
	db, err := store.DB(ctx, "vince")
	if err != nil {
		store.Close()
		return nil, err
	}
	return &Store{store: store, db: db}, nil
}

func (v *Store) engine() *query.LocalEngine {
	return query.NewEngine(memory.DefaultAllocator, v.db.TableProvider())
}

func (v *Store) Session(
	domain string,
	start, end time.Time,
	step time.Duration,
	selectors []*spec.Match,
) {
	filter := []lo.Expr{
		lo.Col("timestamp").Gt(lo.Literal(start.UnixMilli())),
		lo.Col("timestamp").Lt(lo.Literal(end.UnixMilli())),
	}

	for _, s := range selectors {
		filter = append(filter, s.Expr())
	}

	e := v.engine().
		ScanTable(domain)
	e.
		Filter(
			lo.And(filter...),
		).
		Aggregate(
			[]lo.Expr{
				lo.Sum(lo.Col("value")),
				lo.Avg(lo.Col("bounce")),
				lo.Avg(lo.Col("duration")),
			},
			[]lo.Expr{
				lo.Duration(step),
			},
		)
}
