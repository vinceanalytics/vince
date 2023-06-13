package alerts

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Compile transform all ts files to js.
func Compile(dir string) (o map[string][]byte, err error) {
	o = make(map[string][]byte)
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".js" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		o[rel] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return
}
