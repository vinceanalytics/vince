package js

import (
	"errors"
	"io/fs"
	"path/filepath"

	"github.com/evanw/esbuild/pkg/api"
)

type File struct {
	Path string
	Data []byte
}

func Compile(dir string, files []string) ([]*File, error) {
	var scripts []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, e error) error {
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
		scripts = append(scripts, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(scripts) == 0 {
		return []*File{}, nil
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	result := api.Build(api.BuildOptions{
		Target:        api.ES2015,
		EntryPoints:   files,
		Outdir:        dir,
		Outbase:       dir,
		Bundle:        true,
		AbsWorkingDir: dir,
		LogLevel:      api.LogLevelSilent,
	})
	if len(result.Errors) > 0 {
		ls := make([]error, len(result.Errors))
		for k, v := range result.Errors {
			ls[k] = errors.New(v.Text)
		}
		return nil, errors.Join(ls...)
	}
	var o []*File
	for _, v := range result.OutputFiles {
		rel, _ := filepath.Rel(dir, v.Path)
		o = append(o, &File{
			Path: rel,
			Data: v.Contents,
		})
	}
	return o, nil
}
