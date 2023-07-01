package timeseries

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/system"
)

func SaveSystem(ctx context.Context) {
	System(ctx).Read(ctx)
}

type SystemStats struct {
	mu  sync.Mutex
	dir string
	f   *os.File
	ts  time.Time
	w   *parquet.SortingWriter[system.Stats]
}

const parquetFile = "vince.parquet"

func (o *SystemStats) open(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.f != nil {
		err := o.save()
		if err != nil {
			return err
		}
	}
	o.ts = core.Now(ctx)
	var err error
	o.f, err = os.Create(filepath.Join(o.dir, parquetFile))
	if err != nil {
		return err
	}
	if o.w != nil {
		o.w.Reset(o.f)
	} else {
		o.w = parquet.NewSortingWriter[system.Stats](o.f, 10<<10,
			parquet.SortingWriterConfig(
				parquet.SortingColumns(
					parquet.Ascending("timestamp"),
				),
			),
		)
	}
	return nil
}

func (o *SystemStats) save() error {
	f, err := os.Create(filepath.Join(o.dir, ulid.Make().String()))
	if err != nil {
		return err
	}
	err = o.w.Flush()
	if err != nil {
		return err
	}
	_, err = f.ReadFrom(o.f)
	if err != nil {
		return err
	}
	f.Close()
	o.f.Close()
	return nil
}

func openSystem(ctx context.Context, dir string) (*SystemStats, error) {
	os.MkdirAll(dir, 0755)
	var o SystemStats
	o.dir = dir
	return &o, o.open(ctx)
}

func (o *SystemStats) Read(ctx context.Context) {
	o.w.Write([]system.Stats{system.Read(ctx)})
}

func (o *SystemStats) Close() error {
	return o.save()
}
