package alerts

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

// Compile transform all ts files to js.
func Compile(dir string) (scripts []string, err error) {
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".ts" {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "vince.ts" {
			// skip this
			return nil
		}
		scripts = append(scripts, rel)
		return nil
	})
	if err != nil {
		return
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return
	}
	result := api.Build(api.BuildOptions{
		EntryPoints:   scripts,
		Outdir:        dir,
		Bundle:        true,
		Write:         true,
		AbsWorkingDir: dir,
	})
	if len(result.Errors) > 0 {
		ls := make([]error, len(result.Errors))
		for k, v := range result.Errors {
			ls[k] = errors.New(v.Text)
		}
		return nil, errors.Join(ls...)
	}
	return
}
