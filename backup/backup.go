package backup

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/timeseries"
	"github.com/google/uuid"
	"github.com/klauspost/compress/zstd"
	"github.com/mattn/go-sqlite3"
)

const (
	SQLITE_FILE = "sqlite.zstd"
	BADGER_FILE = "badger.zstd"
)

var protect sync.Mutex

// Sqlite creates a full sqlite database compressed with zstd written to w.s
func Sqlite(ctx context.Context, w io.Writer) error {
	cfg := config.Get(ctx)

	path := filepath.Join(cfg.BackupDir, uuid.NewString())
	defer os.Remove(path)

	var driver sqlite3.SQLiteDriver
	d, err := driver.Open(path)
	if err != nil {
		return err
	}
	conn := d.(*sqlite3.SQLiteConn)
	var isClosed bool
	closeConn := func() {
		if !isClosed {
			conn.Close()
			isClosed = true
		}
	}
	defer closeConn()

	db, err := models.Get(ctx).DB()
	if err != nil {
		return err
	}
	bc, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer bc.Close()
	err = bc.Raw(func(driverConn any) error {
		dc := driverConn.(*sqlite3.SQLiteConn)
		b, err := conn.Backup("main", dc, "main")
		if err != nil {
			return err
		}
		defer b.Finish()

		// SQLITE docs says
		//  If N is negative, all remaining source pages are copied
		// we use -1 to copy everything.
		_, err = b.Step(-1)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	closeConn()

	enc, err := zstd.NewWriter(w)
	if err != nil {
		return err
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = enc.ReadFrom(f)
	if err != nil {
		return err
	}
	return enc.Close()
}

// Performs full badger db backup
func Badger(ctx context.Context, w io.Writer) error {
	db := timeseries.GetMike(ctx)
	enc, err := zstd.NewWriter(w)
	if err != nil {
		return err
	}
	_, err = db.Backup(enc, 0)
	if err != nil {
		return err
	}
	return enc.Close()
}

func file(ctx context.Context, w *tar.Writer, name string, b func(context.Context, io.Writer) error) error {
	cfg := config.Get(ctx)
	path := filepath.Join(cfg.BackupDir, uuid.NewString())
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(path)
	}()
	err = b(ctx, f)
	if err != nil {
		return err
	}
	stat, _ := f.Stat()
	err = w.WriteHeader(&tar.Header{
		Name:    name,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	})
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

func Backup(ctx context.Context) error {
	protect.Lock()
	defer protect.Unlock()

	cfg := config.Get(ctx)
	if cfg.BackupDir == "" {
		return nil
	}
	path := filepath.Join(cfg.BackupDir, uuid.NewString())
	defer os.Remove(path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	err = archive(ctx, f)
	if err != nil {
		f.Close()
		return err
	}
	f.Close()
	name := fmt.Sprintf("%s-vince.backup", time.Now().Format(time.DateOnly))
	return os.Rename(path, filepath.Join(cfg.BackupDir, name))
}

func archive(ctx context.Context, w io.Writer) error {
	x := tar.NewWriter(w)
	defer x.Close()
	err := file(ctx, x, SQLITE_FILE, Sqlite)
	if err != nil {
		return err
	}
	return file(ctx, x, BADGER_FILE, Sqlite)
}
