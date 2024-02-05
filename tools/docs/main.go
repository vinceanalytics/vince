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
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Depado/bfchroma/v2"
	bhtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
)

//go:embed page.tmpl
var data string

//go:embed style.css
var styleData []byte

//go:embed reload.js
var reloadData []byte

//go:embed script.js
var scriptData []byte

var style template.CSS
var script []template.JS

var page = template.Must(template.New("main").Parse(data))
var minifier *minify.M

func init() {
	minifier = minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("text/js", js.Minify)
	o, err := minifier.Bytes("text/css", styleData)
	if err != nil {
		panic(err)
	}
	style = template.CSS(o)
	o, err = minifier.Bytes("text/js", scriptData)
	if err != nil {
		panic(err)
	}
	script = append(script, template.JS(o))
}

var serve = flag.Bool("s", true, "serves")

var upgrader = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

const (
	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

const Icon = `href="data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgNTAwIDUwMCIgd2lkdGg9IjUwMCIgaGVpZ2h0PSI1MDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6Yng9Imh0dHBzOi8vYm94eS1zdmcuY29tIj4KICA8cGF0aCBkPSJNIDE2NC4yMDYgMjQ5Ljc5MyBtIC00MS45NDMgMCBhIDQxLjk0MyA0MS45NDMgMCAxIDAgODMuODg2IDAgYSA0MS45NDMgNDEuOTQzIDAgMSAwIC04My44ODYgMCBaIE0gMTY0LjIwNiAyNDkuNzkzIG0gLTI1LjE2NiAwIGEgMjUuMTY2IDI1LjE2NiAwIDAgMSA1MC4zMzIgMCBhIDI1LjE2NiAyNS4xNjYgMCAwIDEgLTUwLjMzMiAwIFoiICB0cmFuc2Zvcm09Im1hdHJpeCgwLjI1NjIzNCwgMC45NjY2MTUsIC0wLjk2NjYxNSwgMC4yNTYyMzQsIDI1MS44NzEwNjMsIDE3MS45NjA4NjEpIiBieDpzaGFwZT0icmluZyAxNjQuMjA2IDI0OS43OTMgMjUuMTY2IDI1LjE2NiA0MS45NDMgNDEuOTQzIDFAZGQ3ODBlYzciLz4KICA8cGF0aCBkPSJNIDE5OS4xMTggMjI3LjY2NCBtIC0xMDEuMzg2IDAgYSAxMDEuMzg2IDEwMS4zODYgMCAxIDAgMjAyLjc3MiAwIGEgMTAxLjM4NiAxMDEuMzg2IDAgMSAwIC0yMDIuNzcyIDAgWiBNIDE5OS4xMTggMjI3LjY2NCBtIC02MC44MzEgMCBhIDYwLjgzMSA2MC44MzEgMCAwIDEgMTIxLjY2MiAwIGEgNjAuODMxIDYwLjgzMSAwIDAgMSAtMTIxLjY2MiAwIFoiICB0cmFuc2Zvcm09Im1hdHJpeCgwLjkzNjQ5LCAwLjM1MDY5NSwgLTAuMzUwNjk1LCAwLjkzNjQ5LCA0MS4yMzc0MzQsIDUxLjIwMDEyNykiIGJ4OnNoYXBlPSJyaW5nIDE5OS4xMTggMjI3LjY2NCA2MC44MzEgNjAuODMxIDEwMS4zODYgMTAxLjM4NiAxQGRiZDI0ZjEyIi8+CiAgPHBhdGggZD0iTSAyMzguNDIxIDE4Mi40NjUgbSAtOTcuNDU4IDAgYSA5Ny40NTggOTcuNDU4IDAgMSAwIDE5NC45MTYgMCBhIDk3LjQ1OCA5Ny40NTggMCAxIDAgLTE5NC45MTYgMCBaIE0gMjM4LjQyMSAxODIuNDY1IG0gLTU4LjQ3MyAwIGEgNTguNDczIDU4LjQ3MyAwIDAgMSAxMTYuOTQ2IDAgYSA1OC40NzMgNTguNDczIDAgMCAxIC0xMTYuOTQ2IDAgWiIgIHRyYW5zZm9ybT0ibWF0cml4KDAuNzczNjA0LCAwLjYzMzY3LCAtMC42MzM2NywgMC43NzM2MDQsIDE4Ni40MTk0NjQsIC04MS40Nzk2ODMpIiBieDpzaGFwZT0icmluZyAyMzguNDIxIDE4Mi40NjUgNTguNDczIDU4LjQ3MyA5Ny40NTggOTcuNDU4IDFAM2ZkMjA3YzkiLz4KICA8cGF0aCBkPSJNIDI3Mi45NjQgMTk5LjAzMSBtIC0xMDAuMzgxIDAgYSAxMDAuMzgxIDEwMC4zODEgMCAxIDAgMjAwLjc2MiAwIGEgMTAwLjM4MSAxMDAuMzgxIDAgMSAwIC0yMDAuNzYyIDAgWiBNIDI3Mi45NjQgMTk5LjAzMSBtIC02MC4yMyAwIGEgNjAuMjMgNjAuMjMgMCAwIDEgMTIwLjQ2IDAgYSA2MC4yMyA2MC4yMyAwIDAgMSAtMTIwLjQ2IDAgWiIgIHRyYW5zZm9ybT0ibWF0cml4KC0wLjk5MzU3OSwgMC4xMTMxNDQsIC0wLjExMzE0NCwgLTAuOTkzNTc5LCA2NDMuMzM2MzY1LCA0MjIuODc4ODQ1KSIgYng6c2hhcGU9InJpbmcgMjcyLjk2NCAxOTkuMDMxIDYwLjIzIDYwLjIzIDEwMC4zODEgMTAwLjM4MSAxQDU2N2I5YzJjIi8+Cjwvc3ZnPg=="`
const LOGO = `src="data:image/svg+xml;base64,PHN2ZyB2aWV3Qm94PSIwIDAgNTAwIDUwMCIgd2lkdGg9IjUwMCIgaGVpZ2h0PSI1MDAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgeG1sbnM6Yng9Imh0dHBzOi8vYm94eS1zdmcuY29tIj4KICA8cGF0aCBkPSJNIDE2NC4yMDYgMjQ5Ljc5MyBtIC00MS45NDMgMCBhIDQxLjk0MyA0MS45NDMgMCAxIDAgODMuODg2IDAgYSA0MS45NDMgNDEuOTQzIDAgMSAwIC04My44ODYgMCBaIE0gMTY0LjIwNiAyNDkuNzkzIG0gLTI1LjE2NiAwIGEgMjUuMTY2IDI1LjE2NiAwIDAgMSA1MC4zMzIgMCBhIDI1LjE2NiAyNS4xNjYgMCAwIDEgLTUwLjMzMiAwIFoiICB0cmFuc2Zvcm09Im1hdHJpeCgwLjI1NjIzNCwgMC45NjY2MTUsIC0wLjk2NjYxNSwgMC4yNTYyMzQsIDI1MS44NzEwNjMsIDE3MS45NjA4NjEpIiBieDpzaGFwZT0icmluZyAxNjQuMjA2IDI0OS43OTMgMjUuMTY2IDI1LjE2NiA0MS45NDMgNDEuOTQzIDFAZGQ3ODBlYzciLz4KICA8cGF0aCBkPSJNIDE5OS4xMTggMjI3LjY2NCBtIC0xMDEuMzg2IDAgYSAxMDEuMzg2IDEwMS4zODYgMCAxIDAgMjAyLjc3MiAwIGEgMTAxLjM4NiAxMDEuMzg2IDAgMSAwIC0yMDIuNzcyIDAgWiBNIDE5OS4xMTggMjI3LjY2NCBtIC02MC44MzEgMCBhIDYwLjgzMSA2MC44MzEgMCAwIDEgMTIxLjY2MiAwIGEgNjAuODMxIDYwLjgzMSAwIDAgMSAtMTIxLjY2MiAwIFoiICB0cmFuc2Zvcm09Im1hdHJpeCgwLjkzNjQ5LCAwLjM1MDY5NSwgLTAuMzUwNjk1LCAwLjkzNjQ5LCA0MS4yMzc0MzQsIDUxLjIwMDEyNykiIGJ4OnNoYXBlPSJyaW5nIDE5OS4xMTggMjI3LjY2NCA2MC44MzEgNjAuODMxIDEwMS4zODYgMTAxLjM4NiAxQGRiZDI0ZjEyIi8+CiAgPHBhdGggZD0iTSAyMzguNDIxIDE4Mi40NjUgbSAtOTcuNDU4IDAgYSA5Ny40NTggOTcuNDU4IDAgMSAwIDE5NC45MTYgMCBhIDk3LjQ1OCA5Ny40NTggMCAxIDAgLTE5NC45MTYgMCBaIE0gMjM4LjQyMSAxODIuNDY1IG0gLTU4LjQ3MyAwIGEgNTguNDczIDU4LjQ3MyAwIDAgMSAxMTYuOTQ2IDAgYSA1OC40NzMgNTguNDczIDAgMCAxIC0xMTYuOTQ2IDAgWiIgIHRyYW5zZm9ybT0ibWF0cml4KDAuNzczNjA0LCAwLjYzMzY3LCAtMC42MzM2NywgMC43NzM2MDQsIDE4Ni40MTk0NjQsIC04MS40Nzk2ODMpIiBieDpzaGFwZT0icmluZyAyMzguNDIxIDE4Mi40NjUgNTguNDczIDU4LjQ3MyA5Ny40NTggOTcuNDU4IDFAM2ZkMjA3YzkiLz4KICA8cGF0aCBkPSJNIDI3Mi45NjQgMTk5LjAzMSBtIC0xMDAuMzgxIDAgYSAxMDAuMzgxIDEwMC4zODEgMCAxIDAgMjAwLjc2MiAwIGEgMTAwLjM4MSAxMDAuMzgxIDAgMSAwIC0yMDAuNzYyIDAgWiBNIDI3Mi45NjQgMTk5LjAzMSBtIC02MC4yMyAwIGEgNjAuMjMgNjAuMjMgMCAwIDEgMTIwLjQ2IDAgYSA2MC4yMyA2MC4yMyAwIDAgMSAtMTIwLjQ2IDAgWiIgIHRyYW5zZm9ybT0ibWF0cml4KC0wLjk5MzU3OSwgMC4xMTMxNDQsIC0wLjExMzE0NCwgLTAuOTkzNTc5LCA2NDMuMzM2MzY1LCA0MjIuODc4ODQ1KSIgYng6c2hhcGU9InJpbmcgMjcyLjk2NCAxOTkuMDMxIDYwLjIzIDYwLjIzIDEwMC4zODEgMTAwLjM4MSAxQDU2N2I5YzJjIi8+Cjwvc3ZnPg=="`

func main() {
	flag.Parse()

	if *serve {
		script = append(script, template.JS(reloadData))
		reload := make(chan struct{})
		var b bytes.Buffer
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					log.Println("event:", event)
					reload <- struct{}{}
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				}
			}
		}()
		err = watcher.Add(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.URL.Path)
			switch r.URL.Path {
			case "/":
				b.Reset()
				err := Build(&b, flag.Arg(0))
				if err != nil {
					log.Println("Build:", err)
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write(b.Bytes())
				return
			case "/reload":
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					log.Println("Upgrade:", err)
					return
				}
				pingTicker := time.NewTicker(pingPeriod)
				defer func() {
					pingTicker.Stop()
					conn.Close()
				}()
				for {
					select {
					case <-reload:
						w, err := conn.NextWriter(websocket.TextMessage)
						if err != nil {
							return
						}
						w.Write([]byte("reload"))
					case <-pingTicker.C:
						if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
							return
						}
					}
				}
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		})
		println("http://localhost:8081")
		err = http.ListenAndServe(":8081", h)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
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
		CSS:  style,
		JS:   script,
		Logo: LOGO,
		Icon: Icon,
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
				lastNode.HeadingData.HeadingID = toLink(id, string(node.Literal))
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

var seen = map[string]struct{}{}

func toLink(id func() int, txt string) string {
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
	CSS   template.CSS
	Logo  template.HTMLAttr
	Icon  template.HTMLAttr
	JS    []template.JS
	Menus []Menu
	Pages []template.HTML
}

type Menu struct {
	ID    string
	Text  string
	Items []Item
}
