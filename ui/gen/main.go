package main

import (
	"io/fs"
	"path/filepath"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	tools.ExecPlain("npm", "run", "build")
	tools.Remove("../assets/ui/")
	tools.CopyDir("build/", "../assets/ui/")
	tools.Remove("../assets/ui/asset-manifest.json")
	monaco("loader.js")
	monaco("editor", "editor.main.js")
	monaco("editor", "editor.main.nls.js")
	monaco("editor", "editor.main.css")
	for _, b := range base() {
		monaco(b...)
	}
	for _, b := range languages() {
		monaco(b...)
	}
	maps()
}

func monaco(src ...string) {
	dst := filepath.Join(append([]string{"../assets", "ui", "vs"}, src...)...)
	tools.MkDir(filepath.Dir(dst))
	tools.Copy(
		dst,
		append([]string{"node_modules", "monaco-editor", "min", "vs"}, src...)...,
	)
}

func base() (o [][]string) {
	vs := filepath.Join("node_modules", "monaco-editor", "min", "vs")
	root := filepath.Join(vs, "base")
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".js", ".css", ".ttf":
		default:
			return nil
		}
		r, _ := filepath.Rel(vs, path)
		o = append(o, filepath.SplitList(r))
		return nil
	})
	return
}

func languages() (o [][]string) {
	vs := filepath.Join("node_modules", "monaco-editor", "min", "vs")
	root := filepath.Join(vs, "basic-languages")
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch filepath.Base(path) {
		case "mysql.js":
		default:
			return nil
		}
		r, _ := filepath.Rel(vs, path)
		o = append(o, filepath.SplitList(r))
		return nil
	})
	return
}

func maps() {
	vs := filepath.Join("node_modules", "monaco-editor")
	root := filepath.Join(vs, "min-maps", "vs")
	dst := []string{"../assets", "ui"}
	filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		r, _ := filepath.Rel(vs, path)
		if info.IsDir() {
			tools.MkDir(append(dst, r)...)
			return nil
		}
		tools.Copy(
			filepath.Join(append(dst, r)...),
			path,
		)
		return nil
	})
}
