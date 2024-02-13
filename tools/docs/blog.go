package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/Depado/bfchroma/v2"
	bhtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gosimple/slug"
	"github.com/russross/blackfriday/v2"
)

//go:embed post.tmpl
var postData string

var post = template.Must(template.New("main").Parse(postData))

type Blog struct{}

type BlogSection struct {
	URL   string
	Title string
	Posts []*Post
}

type Post struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    string
	Section   string
	Title     string
	URL       string
	Content   string
}

const (
	createdAt = "createdAt"
	updatedAt = "updatedAt"
	author    = "author"
	section   = "section"
)

func renderPost(text []byte) (o Post) {
	w := new(bytes.Buffer)
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
			case 1:
				o.Title = string(node.Literal)
				o.URL = slug.Make(o.Title)
			case 6:
				txt := string(node.Literal)
				key, value, _ := strings.Cut(txt, " ")
				switch key {
				case createdAt:
					ts, err := time.Parse(time.DateOnly, value)
					if err != nil {
						fail(err)
					}
					o.CreatedAt = ts
				case updatedAt:
					ts, err := time.Parse(time.DateOnly, value)
					if err != nil {
						fail(err)
					}
					o.UpdatedAt = ts
				case author:
					o.Author = value
				case section:
					o.Author = section
				}

			}
		}
		return blackfriday.GoToNext
	})

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return r.RenderNode(w, node, entering)
	})
	o.Content = w.String()
	return
}

func fail(err error) {
	log.Fatal(err)
}
