package timeseries

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/gernest/vince/log"
	"github.com/gernest/vince/system"
	"github.com/oklog/ulid/v2"
	"github.com/segmentio/parquet-go"
)

type System[T any] struct {
	mu        sync.Mutex
	f         *os.File
	dir, name string
	w         *parquet.SortingWriter[T]
}

func NewSystem[T any](dir, name string) (*System[T], error) {
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	s := &System[T]{
		f:    f,
		dir:  dir,
		name: name,
		w: parquet.NewSortingWriter[T](f, 64<<10,
			parquet.SortingWriterConfig(
				parquet.SortingColumns(
					parquet.Descending("timestamp"),
				),
			),
		),
	}
	return s, nil

}

func (s *System[T]) Write(rows []T) (int, error) {
	return s.w.Write(rows)
}

func (s *System[T]) Save(reopen bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.w.Close()
	if err != nil {
		return err
	}
	s.f.Close()
	err = os.Rename(
		filepath.Join(s.dir, s.name),
		filepath.Join(s.dir, "metrics", s.name+"-"+ulid.Make().String()),
	)
	if err != nil {
		return err
	}
	if reopen {
		s.f, err = os.Create(filepath.Join(s.dir, s.name))
		if err != nil {
			return err
		}
		s.w.Reset(s.f)
	}
	return nil
}

func (s *System[T]) Close() error {
	return s.Save(false)
}

type AllSystem struct {
	system    *System[*system.Stats]
	histogram *System[*system.Histogram]
}

func openSystem(dataPath string) (*AllSystem, error) {
	path := filepath.Join(dataPath, "system")
	os.MkdirAll(filepath.Join(path, "metrics"), 0755)
	c, err := NewSystem[*system.Stats](path, "stats")
	if err != nil {
		return nil, err
	}
	h, err := NewSystem[*system.Histogram](path, "histogram")
	if err != nil {
		return nil, err
	}
	return &AllSystem{system: c, histogram: h}, nil
}

func (a *AllSystem) Close() error {
	return errors.Join(
		a.system.Close(), a.histogram.Close(),
	)
}

func (a *AllSystem) Save() error {
	return errors.Join(
		a.system.Save(true), a.histogram.Save(true),
	)
}

func (a *AllSystem) Collect(ctx context.Context) func(s system.Stats, h ...*system.Histogram) {
	var ls [1]*system.Stats
	return func(s system.Stats, h ...*system.Histogram) {
		ls[0] = &s
		_, err := a.system.Write(ls[:])
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save system stats")
		}
		_, err = a.histogram.Write(h)
		if err != nil {
			log.Get(ctx).Err(err).Msg("failed to save histogram stats ")
		}
	}
}
