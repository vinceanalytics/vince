package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/Depado/bfchroma/v2"
	bhtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/russross/blackfriday/v2"
)

func writeDocs(src, out string) error {
	b := new(bytes.Buffer)
	err := buildDocs(b, src)
	if err != nil {
		return fmt.Errorf("failed building documentation %v", err)
	}
	o, err := minifier.Bytes("text/html", b.Bytes())
	if err != nil {
		return fmt.Errorf("failed minifying documentation %v", err)
	}
	return os.WriteFile(filepath.Join(out, "index.html"), o, 0600)
}

func buildDocs(w io.Writer, dir string) error {
	var idx int
	id := func() int {
		idx++
		return idx
	}
	var b bytes.Buffer
	m := Model{
		Domain: domain,
		Track:  track,
		CSS:    style,
		JS:     script,
		Logo:   LOGO,
		Icon:   Icon,
	}
	var positions []int
	seen := make(map[string]struct{})
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if filepath.Ext(path) != ".md" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		b.Reset()
		items := renderPage(&b, id, seen, data)
		name := filepath.Base(path)
		order, name, _ := strings.Cut(name, "-")
		i, err := strconv.Atoi(order)
		if err != nil {
			return err
		}
		positions = append(positions, i)
		x := Menu{
			ID:    items[0].ID,
			Text:  items[0].Text,
			Items: items[1:],
		}
		if len(items) > 0 {
			x.ID = items[0].ID
		}
		m.Menus = append(m.Menus, x)
		m.Pages = append(m.Pages, template.HTML(b.String()))
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

func renderPage(w io.Writer, id func() int, seen map[string]struct{}, text []byte) (o []Item) {
	m := blackfriday.New(
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
	)
	r := &bfchroma.Renderer{
		Base: blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			Flags: blackfriday.CommonHTMLFlags,
		}),
		Style:      styles.SolarizedLight,
		Autodetect: true,
	}
	r.Formatter = bhtml.New(r.ChromaOptions...)
	ast := m.Parse(text)
	var inHeading bool
	var lastNode *blackfriday.Node
	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Heading && !node.HeadingData.IsTitleblock {
			inHeading = entering
			if entering {
				lastNode = node
			}
			return blackfriday.GoToNext
		}
		if inHeading {
			switch lastNode.HeadingData.Level {
			case 1, 2:
				lastNode.HeadingData.HeadingID = toLink(id, seen, string(node.Literal))
				o = append(o, Item{
					ID:   lastNode.HeadingData.HeadingID,
					Text: string(node.Literal),
				})
			}
		}
		return blackfriday.GoToNext
	})

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(w, node, entering)
	})
	return
}

func toLink(id func() int, seen map[string]struct{}, txt string) string {
	txt = strings.Replace(txt, " ", "-", -1)
	txt = strings.ToLower(txt)
	_, ok := seen[txt]
	if !ok {
		seen[txt] = struct{}{}
		return txt
	}
	return fmt.Sprintf("%d-%s", id(), txt)
}

type Item struct {
	ID   string
	Text string
}

type Model struct {
	Domain string
	Track  string
	CSS    template.CSS
	Logo   template.HTMLAttr
	Icon   template.HTMLAttr
	JS     []template.JS
	Menus  []Menu
	Pages  []template.HTML
}

type Menu struct {
	ID    string
	Text  string
	Items []Item
}
