package log

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Rotate struct {
	filename string
	f        *os.File
	w        *bufio.Writer
	mu       sync.Mutex
}

func NewRotate(path string) (*Rotate, error) {
	b := &Rotate{
		filename: filepath.Join(path, "error_log"),
	}
	return b, b.open()
}

func (b *Rotate) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.w.Write(p)
}

func (b *Rotate) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.w.Flush()
	return b.f.Sync()
}

func (b *Rotate) Rotate() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	err := b.close()
	if err != nil {
		return err
	}
	return b.reopen()
}

func (b *Rotate) open() error {
	var err error
	b.f, err = os.Create(b.filename)
	if err != nil {
		return fmt.Errorf("failed to create new log file %s  %v", b.filename, err)
	}
	b.w = bufio.NewWriter(b.f)
	return nil
}

func (b *Rotate) reopen() error {
	var err error
	b.f, err = os.Create(b.filename)
	if err != nil {
		return fmt.Errorf("failed to create new log file %s  %v", b.filename, err)
	}
	b.w.Reset(b.f)
	return nil
}

func (b *Rotate) Close() error {
	return b.close()
}

func (b *Rotate) close() error {
	err := b.w.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush messages %v", err)
	}
	now := time.Now().UTC().Format(time.DateOnly)
	out := filepath.Join(filepath.Dir(b.filename), now+".log.gz")
	o, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("failed to create rotation file %s  %v", out, err)
	}
	defer o.Close()

	_, err = b.f.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to seek log file %v", err)
	}
	w := gzip.NewWriter(o)
	_, err = io.Copy(w, b.f)
	if err != nil {
		return fmt.Errorf("failed to copy rotation file %v", err)
	}
	w.Close()
	err = b.f.Close()
	if err != nil {
		return fmt.Errorf("failed to close log file writer %v", err)
	}
	return nil
}
