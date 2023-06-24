package js

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/email"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/packages"
)

type file struct {
	Path string
	Data []byte
}

const vinceFile = "__vince__.ts"

type Alert struct {
	Name     string
	Path     string
	Interval time.Duration
	VM       *goja.Runtime
	JS       []byte
	Function goja.Callable
}

func (a *Alert) Run(ctx context.Context) {

}

func (a *Alert) schedule(call goja.Callable) {
	a.Function = call
}

var relVince = []byte("./" + vinceFile)
var vinceImport = []byte("@vinceanalytics/vince")

func Compile(ctx context.Context, paths ...string) ([]*Alert, error) {
	dir, err := os.MkdirTemp("", "vince_alerts")
	if err != nil {
		return nil, err
	}
	defer func() {
		os.Remove(dir)
	}()
	return CompileWith(ctx, dir, paths...)
}

func CompileWith(ctx context.Context, dir string, paths ...string) ([]*Alert, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	vs := filepath.Join(dir, vinceFile)
	err = os.WriteFile(vs, packages.VINCE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write __vince__ file %v", err)
	}
	namer := make(map[string]*Alert)
	var i uint
	scripts := []string{}
	for _, p := range paths {
		name, path, interval, err := config.ParseAlert(p)
		if err != nil {
			return nil, err
		}
		x := &Alert{
			Name:     name,
			Path:     path,
			Interval: interval,
			VM:       goja.New(),
		}
		x.VM.Set("__schedule__", x.schedule)
		r, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		r = bytes.ReplaceAll(r, vinceImport, relVince)
		o := filepath.Base(path)
		if ext := filepath.Ext(o); ext != ".js" {
			// a typescript file
			xo := strings.ReplaceAll(o, ext, ".js")
			if _, ok := namer[xo]; ok {
				// We have already copied file with the same base name
				i++
				o = fmt.Sprintf("%d%s", i, o)
				xo = fmt.Sprintf("%d%s", i, xo)
			}
			namer[xo] = x
		} else {
			if _, ok := namer[o]; ok {
				// We have already copied file with the same base name
				i++
				o = fmt.Sprintf("%d%s", i, o)
			}
			namer[o] = x
		}
		s := filepath.Join(dir, o)
		err = os.WriteFile(s, r, 0600)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, s)
	}

	result := api.Build(api.BuildOptions{
		EntryPoints:    scripts,
		Outdir:         dir,
		Outbase:        dir,
		AbsWorkingDir:  dir,
		Write:          true,
		AllowOverwrite: true,
		Bundle:         true,
		LogLevel:       api.LogLevelSilent,
	})
	if len(result.Errors) > 0 {
		ls := make([]error, len(result.Errors))
		for k, v := range result.Errors {
			ls[k] = errors.New(v.Text)
		}
		return nil, errors.Join(ls...)
	}
	var o []*Alert
	for _, v := range result.OutputFiles {
		rel, _ := filepath.Rel(dir, v.Path)
		base := filepath.Base(rel)
		a, ok := namer[base]
		if !ok {
			continue
		}
		a.JS = v.Contents

		err = load(ctx, a.VM, a.JS)
		if err != nil {
			return nil, err
		}
		o = append(o, a)
	}
	sort.Slice(o, func(i, j int) bool {
		return o[i].Name < o[j].Name
	})
	return o, nil
}

func load(ctx context.Context, vm *goja.Runtime, mainPkg []byte) error {
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	query.Register(vm)
	email.Register(ctx, vm)
	vm.Set("__query__", queryStats(ctx))
	_, err := vm.RunString(string(mainPkg))
	return err
}
