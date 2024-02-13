package main

import (
	"bytes"
	_ "embed"
	"html/template"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
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

type Blog struct {
	Title  string
	Domain string
	Track  string
	CSS    template.CSS
	Logo   template.HTMLAttr
	Icon   template.HTMLAttr
	JS     []template.JS

	Post     *Post
	Sections Sections
	Section  *Section
}

type Section struct {
	URL       string
	Title     string
	Posts     Posts
	Timestamp int64
}

func (s *Section) Update() {
	for i := range s.Posts {
		s.Posts[i].URL = path.Join(s.URL, s.Posts[i].URL)
	}
}

func writeBlogFile(path string, ctx Blog) error {
	b := new(bytes.Buffer)
	err := post.Execute(b, ctx)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "index.html"), b.Bytes(), 0600)
}

func (s *Section) Write(base string) error {
	for _, p := range s.Posts {
		err := os.MkdirAll(filepath.Join(base, p.URL), 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
		err = p.Write(base)
		if err != nil {
			return err
		}
	}
	return writeBlogFile(filepath.Join(base, s.URL), Blog{
		Title:   s.Title,
		Section: s,

		Domain: domain,
		Track:  track,
		CSS:    style,
		JS:     script,
		Logo:   LOGO,
		Icon:   Icon,
	})
}

func writeBlog(src, out string) error {
	src = filepath.Join(src, "blog")
	out = filepath.Join(out, "blog")
	os.MkdirAll(out, 0755)
	dir, err := os.ReadDir(src)
	if err != nil {
		if os.IsNotExist(err) {
			println("no blog directory skipping blog generation")
			return nil
		}
		return err
	}

	var posts Posts
	for _, e := range dir {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".md" {
			continue
		}
		d, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		posts = append(posts, renderPost(d))
	}
	sections := posts.Sections()
	return sections.Write(out)
}

type Sections []*Section

func (ls Sections) Update() {
	for _, s := range ls {
		s.Update()
	}
}

func (ls Sections) Write(base string) error {
	ls.Update()
	for _, s := range ls {
		err := os.MkdirAll(filepath.Join(base, s.URL), 0755)
		if err != nil && !os.IsExist(err) {
			return err
		}
		err = s.Write(base)
		if err != nil {
			return err
		}
	}
	return writeBlogFile(base, Blog{
		Title:    "vince- Blog",
		Sections: ls,

		Domain: domain,
		Track:  track,
		CSS:    style,
		JS:     script,
		Logo:   LOGO,
		Icon:   Icon,
	})
}

func (ls Sections) Len() int {
	return len(ls)
}

func (ls Sections) Less(i, j int) bool {
	return ls[i].Timestamp < ls[j].Timestamp
}

func (ls Sections) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

type Post struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    string
	Section   string
	Title     string
	URL       string
	Content   template.HTML
}

func (p *Post) Date() string {
	return time.UnixMilli(p.Timestamp()).Format(time.DateOnly)
}

func (p *Post) Write(base string) error {
	return writeBlogFile(filepath.Join(base, p.URL), Blog{
		Title: p.Title,
		Post:  p,

		Domain: domain,
		Track:  track,
		CSS:    style,
		JS:     script,
		Logo:   LOGO,
		Icon:   Icon,
	})
}

func (p *Post) Timestamp() int64 {
	if p.UpdatedAt.After(p.CreatedAt) {
		return p.UpdatedAt.UnixMilli()
	}
	return p.CreatedAt.UnixMilli()
}

type Posts []Post

func (ls Posts) Len() int {
	return len(ls)
}

func (ls Posts) Less(i, j int) bool {
	return ls[i].Timestamp() < ls[j].Timestamp()
}

func (ls Posts) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

func (ls Posts) Timestamp() (o int64) {
	for i, p := range ls {
		if i == 0 {
			o = p.Timestamp()
			continue
		}
		o = max(o, p.Timestamp())
	}
	return
}

func (ls Posts) Sections() (o Sections) {
	m := make(map[string]Posts)
	for _, p := range ls {
		m[p.Section] = append(m[p.Section], p)
	}
	for k, v := range m {
		sort.Sort(v)
		o = append(o, &Section{
			Title:     k,
			URL:       slug.Make(k),
			Posts:     v,
			Timestamp: v.Timestamp(),
		})
	}
	sort.Sort(o)
	return
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
					o.Section = value
				}
			}
		}
		return blackfriday.GoToNext
	})

	ast.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		if node.Type == blackfriday.Heading && node.HeadingData.Level == 6 {
			return blackfriday.SkipChildren
		}
		return r.RenderNode(w, node, entering)
	})
	o.Content = template.HTML(w.String())
	return
}

func fail(err error) {
	log.Fatal(err)
}
