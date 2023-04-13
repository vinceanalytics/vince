package main

import (
	"bytes"
	"flag"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gernest/vince/assets/ui/templates"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
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

	var buf bytes.Buffer
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
		buf.Reset()
		w := m.Writer("text/html", &buf)
		err = Render(w, b, filepath.Base(path), info.ModTime())
		if err != nil {
			return err
		}
		path = strings.TrimSuffix(path, filepath.Ext(path))
		path += ".html"
		path, _ = filepath.Rel(src, path)
		w.Close()
		return os.WriteFile(filepath.Join(dest, path), buf.Bytes(), info.Mode())
	})
	if err != nil {
		log.Fatal(err)
	}
}

func Render(w io.Writer, b []byte, title string, mod time.Time) error {
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(b)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	b = markdown.Render(doc, renderer)
	return templates.Markdown.Execute(w, &templates.Context{
		Title:   title,
		Content: template.HTML(b),
		ModTime: mod,
		Docs:    true,
	})
}
