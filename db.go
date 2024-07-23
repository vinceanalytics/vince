package len64

import (
	"context"
	"log/slog"
	"time"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
)

const (
	timestampField = "_timestamp"
	dateField      = "_date"
)

type DB struct {
	store *Store[*v1.Model]
	tasks chan *task
}

func (db *DB) Start(ctx context.Context) error {
	b, err := db.store.Batch()
	if err != nil {
		return err
	}
	go db.startBatch(b, ctx)
	return nil
}

func (db *DB) startBatch(b *Batch[*v1.Model], ctx context.Context) {
	ts := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-db.tasks:
			err := b.Write(e.model, func(idx Index) {
				for k, v := range e.meta {
					idx.String(k, v)
				}
				idx.Int64(timestampField, int64(e.model.Timestamp))
				idx.Int64(dateField, date(e.model.Timestamp))
			})
			if err != nil {
				slog.Error("writing model", "err", err)
			}
		case <-ts.C:
			err := b.Flush()
			if err != nil {
				slog.Error("flushing batch", "err", err)
			}
		}
	}
}

func (db *DB) Save(model *v1.Model, metadata map[string]string) {
	db.tasks <- &task{model: model, meta: metadata}
}

func date(ts uint64) int64 {
	yy, mm, dd := time.UnixMilli(int64(ts)).Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).UnixMilli()
}

type task struct {
	model *v1.Model
	meta  map[string]string
}
