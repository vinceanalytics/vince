package len64

import (
	"context"
	"log/slog"
	"time"

	v1 "github.com/gernest/len64/gen/go/len64/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type DB struct {
	store *Store
	tasks chan *v1.Model
}

func Open(path string) (*DB, error) {
	db, err := newStore(path, false)
	if err != nil {
		return nil, err
	}
	return &DB{store: db, tasks: make(chan *v1.Model, 1<<10)}, nil
}

func (db *DB) Close() error {
	close(db.tasks)
	return db.store.Close()
}

func (db *DB) Start(ctx context.Context) error {
	b, err := db.store.Batch()
	if err != nil {
		return err
	}
	go db.startBatch(b, ctx)
	return nil
}

func (db *DB) startBatch(b *Batch, ctx context.Context) {
	ts := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-db.tasks:
			err := b.Write(e.Timestamp, func(idx Index) {
				e.ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
					if fd.Kind() == protoreflect.StringKind {
						idx.String(string(fd.Name()), v.String())
					}
					return true
				})
				idx.Int64("timestamp", int64(e.Timestamp))
				idx.Int64("date", date(e.Timestamp))
				idx.Int64("uid", int64(e.Id))
				idx.Bool("bounce", e.Bounce)
				idx.Bool("session", e.Session)
				idx.Int64("duration", int64(e.Duration))
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

func (db *DB) Save(model *v1.Model) {
	db.tasks <- model
}

func date(ts uint64) int64 {
	yy, mm, dd := time.UnixMilli(int64(ts)).Date()
	return time.Date(yy, mm, dd, 0, 0, 0, 0, time.UTC).UnixMilli()
}
