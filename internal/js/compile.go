package js

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/vinceanalytics/vince/packages"
)

type file struct {
	Path string
	Data []byte
}

const vinceFile = "__vince__.ts"

func Compile(ctx context.Context, dir string) (*File, error) {
	err := os.WriteFile(filepath.Join(dir, vinceFile), packages.VINCE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write __vince__ file %v", err)
	}
	defer func() {
		os.Remove(filepath.Join(dir, vinceFile))
	}()
	var scripts []string
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, e error) error {
		if e != nil {
			return e
		}
		if info.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".ts", ".js":
		default:
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
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	result := api.Build(api.BuildOptions{
		EntryPoints:   scripts,
		Outdir:        dir,
		Outbase:       dir,
		AbsWorkingDir: dir,
		Format:        api.FormatCommonJS,
		LogLevel:      api.LogLevelSilent,
	})
	if len(result.Errors) > 0 {
		ls := make([]error, len(result.Errors))
		for k, v := range result.Errors {
			ls[k] = errors.New(v.Text)
		}
		return nil, errors.Join(ls...)
	}
	var o []*file
	var vincePkg []byte
	for _, v := range result.OutputFiles {
		rel, _ := filepath.Rel(dir, v.Path)
		base := filepath.Base(rel)
		if base == vinceFile {
			vincePkg = v.Contents
			continue
		}
		o = append(o, &file{
			Path: rel,
			Data: v.Contents,
		})
	}
	vm := create(ctx)
	pkg := vm.runtime.NewObject()
	vm.runtime.Set("module", pkg)
	_, err = vm.runtime.RunString(string(vincePkg))
	if err != nil {
		return nil, err
	}

	vm.runtime.Set("require", func(a goja.Value) goja.Value {
		if a.String() == "@vinceanalytics/vince" {
			return pkg.Get("exports")
		}
		return goja.Undefined()
	})

	for _, m := range o {
		_, err = vm.runtime.RunScript(m.Path, string(m.Data))
		if err != nil {
			return nil, err
		}
	}
	return vm, nil
}
