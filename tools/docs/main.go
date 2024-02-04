package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Depado/bfchroma/v2"
	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
)

//go:embed page.tmpl
var data string

//go:embed style.css
var styleData []byte

var style template.CSS

var page = template.Must(template.New("main").Parse(data))
var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/html", html.Minify)
	o, err := minifier.Bytes("text/css", styleData)
	if err != nil {
		panic(err)
	}
	style = template.CSS(o)
}

func main() {
	flag.Parse()
	var b bytes.Buffer
	err := Build(&b, flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	o, err := minifier.Bytes("text/html", b.Bytes())
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(flag.Arg(1), o, 0600)
	if err != nil {
		log.Fatal(err)
	}
}

func Build(w io.Writer, dir string) error {
	var idx int
	id := func() int {
		idx++
		return idx
	}
	var b bytes.Buffer
	m := Model{
		CSS: style,
	}
	var positions []int
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b.Reset()
		items := renderPage(&b, id, data)
		name := filepath.Base(path)
		order, name, _ := strings.Cut(name, "-")
		i, err := strconv.Atoi(order)
		if err != nil {
			return err
		}
		positions = append(positions, i)
		name = strings.TrimSuffix(name, filepath.Ext(path))
		name = strings.ReplaceAll(name, "-", " ")
		x := Menu{
			Text:  name,
			Items: items,
		}
		if len(items) > 0 {
			x.ID = items[0].ID
		}
		m.Menus = append(m.Menus, x)
		m.Pages = append(m.Pages, template.HTML(b.String()))
		return nil
	})
	if err != nil {
		return err
	}
	x := &ms{indices: positions, m: &m}
	sort.Sort(x)
	return page.Execute(w, m)
}

type ms struct {
	indices []int
	m       *Model
}

var _ sort.Interface = (*ms)(nil)

func (m *ms) Len() int {
	return len(m.indices)
}

func (m *ms) Less(i, j int) bool {
	return m.indices[i] < m.indices[j]
}

func (m *ms) Swap(i, j int) {
	m.indices[i], m.indices[j] = m.indices[j], m.indices[i]
	m.m.Menus[i], m.m.Menus[j] = m.m.Menus[j], m.m.Menus[i]
	m.m.Pages[i], m.m.Pages[j] = m.m.Pages[j], m.m.Pages[i]
}

func renderPage(w io.Writer, id func() int, text []byte) (o []Item) {
	m := blackfriday.New(
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
	)

	r := bfchroma.NewRenderer()
	ast := m.Parse(text)
	var inHeading bool
	var level int
	var count string

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Heading && !node.HeadingData.IsTitleblock {
			inHeading = entering
			if entering {
				node.HeadingData.HeadingID = strconv.Itoa(id())
				level = node.HeadingData.Level
				count = node.HeadingID
			}
			return blackfriday.GoToNext
		}
		if inHeading {
			if level == 2 {
				// only count top level
				o = append(o, Item{
					ID:   count,
					Text: string(node.Literal),
				})
				fmt.Println(level, count, string(node.Literal))
			}
		}
		return blackfriday.GoToNext
	})
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(w, node, entering)
	})
	return
}

type Item struct {
	ID   string
	Text string
}

type Model struct {
	CSS   template.CSS
	JS    template.JS
	Menus []Menu
	Pages []template.HTML
}

type Menu struct {
	ID    string
	Text  string
	Items []Item
}
