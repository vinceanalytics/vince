package timeseries

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/segmentio/parquet-go"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/system"
	"github.com/vinceanalytics/vince/pkg/log"
	"github.com/vinceanalytics/vince/pkg/spec"
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

var schema = parquet.SchemaOf(system.Stats{})

var SystemFields = fields()

func fields() (f []string) {
	for _, e := range schema.Fields()[1:] {
		f = append(f, e.Name())
	}
	return
}
func (o *SystemStats) Query(ctx context.Context, paths ...string) (r spec.System) {
	leaf, ok := schema.Lookup(paths...)
	if !ok {
		return
	}
	if time.Since(o.ts) > (15 * time.Minute) {
		o.mu.Lock()
		o.save()
		o.mu.Unlock()
	}
	err := o.read(o.readColum(leaf, &r))
	if err != nil {
		log.Get().Err(err).Msg("failed querying system stats")
	}
	return
}

func (o *SystemStats) readColum(col parquet.LeafColumn, rs *spec.System) func(io.ReaderAt, int64) error {
	return func(ra io.ReaderAt, i int64) error {
		f, err := parquet.OpenFile(ra, i)
		if err != nil {
			return err
		}
		for _, g := range f.RowGroups() {
			chunks := g.ColumnChunks()
			ts, err := readInt64(chunks[0])
			if err != nil {
				return err
			}
			rs.Timestamps = append(rs.Timestamps, ts...)
			column, err := readInt64(chunks[col.ColumnIndex])
			if err != nil {
				return err
			}
			rs.Result = append(rs.Result, column...)
		}
		return nil
	}
}

func readInt64(col parquet.ColumnChunk) ([]int64, error) {
	pages := col.Pages()
	defer pages.Close()
	var result []int64
	for {
		p, err := pages.ReadPage()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return result, nil
			}
			return nil, err
		}
		switch e := p.Values().(type) {
		case parquet.Int64Reader:
			o := make([]int64, p.NumValues())
			_, err := e.ReadInt64s(o)
			if err != nil {
				return nil, err
			}
			result = append(result, o...)
		}
	}
}

func (o *SystemStats) read(fn func(io.ReaderAt, int64) error) error {
	var files []ulid.ULID
	fi, err := os.ReadDir(o.dir)
	if err != nil {
		return err
	}
	for _, f := range fi {
		u, err := ulid.Parse(f.Name())
		if err != nil {
			continue
		}
		files = append(files, u)
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Compare(files[i]) == -1
	})
	for _, u := range files {
		f, err := os.Open(filepath.Join(o.dir, u.String()))
		if err != nil {
			return err
		}
		stat, err := f.Stat()
		if err != nil {
			f.Close()
			return err
		}
		err = fn(f, stat.Size())
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
