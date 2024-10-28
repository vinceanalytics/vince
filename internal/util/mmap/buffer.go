package mmap

import (
	"os"
)

type Source struct {
	file *MmapFile
}

func NewSource(path string, capacity int) (*Source, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	mmapFile, err := OpenMmapFileUsing(file, capacity, true)
	if err != nil && err != NewFile {
		return nil, err
	}
	return &Source{file: mmapFile}, nil
}

func (s *Source) Data() []byte {
	return s.file.Data
}

func (s *Source) Grow(capacity int) []byte {
	err := s.file.Truncate(int64(capacity))
	if err != nil {
		panic("mmap: truncating file")
	}
	return s.file.Data
}

func (s *Source) Release() error {
	return s.file.Close(-1)
}
