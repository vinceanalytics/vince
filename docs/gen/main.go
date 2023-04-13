package main

import (
	"bytes"
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	h2 "github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/svg"
)

func main() {
	flag.Parse()
	src := flag.Arg(0)
	if src == "" {
		return
	}
	dest := flag.Arg(1)
	if dest == "" {
		return
	}
	dest, err := filepath.Abs(dest)
	if err != nil {
		log.Fatal(err)
	}
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", h2.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)

	var pages templates.Pages
	err = filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			path, _ = filepath.Rel(src, path)
			if path != "" {
				os.Mkdir(filepath.Join(dest, path), info.Mode())
			}
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		path = strings.TrimSuffix(path, filepath.Ext(path))
		path += ".html"
		path, _ = filepath.Rel(src, path)
		var p templates.Page
		p.Read(path, b, info.ModTime())
		pages = append(pages, &p)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	pages.Sort()
	var b bytes.Buffer
	for _, p := range pages {
		b.Reset()
		w := m.Writer("text/html", &b)
		err := p.Render(w, pages)
		if err != nil {
			log.Fatal("failed to render page", err)
		}
		w.Close()
		err = os.WriteFile(filepath.Join(dest, p.Path), b.Bytes(), 0600)
		if err != nil {
			log.Fatal("failed to write page", err)
		}
	}
}
