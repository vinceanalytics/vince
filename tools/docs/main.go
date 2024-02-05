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
	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/russross/blackfriday/v2"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
)

//go:embed page.tmpl
var data string

//go:embed style.css
var styleData []byte

//go:embed reload.js
var reloadData []byte

var style template.CSS
var script []template.JS

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
		CSS: style,
		JS:  script,
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

	r := bfchroma.NewRenderer()
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
	JS    []template.JS
	Menus []Menu
	Pages []template.HTML
}

type Menu struct {
	ID    string
	Text  string
	Items []Item
}
