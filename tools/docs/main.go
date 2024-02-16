package main

import (
	_ "embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
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

var domain = os.Getenv("DOMAIN")
var track = os.Getenv("TRACKER")

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

	if v := os.Getenv("TRACKER"); v != "" {
		track = v
	}
	if v := os.Getenv("DOMAIN"); v != "" {
		domain = v
	}
}

var serve = flag.Bool("s", false, "serves")

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
	src := flag.Arg(0)
	out := flag.Arg(1)
	if *serve {
		out, err := os.MkdirTemp("", "docs")
		if err != nil {
			fail(err)
		}
		defer os.RemoveAll(out)
		println("serving from", out)
		script = append(script, template.JS(reloadData))
		err = build(src, out)
		if err != nil {
			fail(err)
		}
		fsv := http.FileServerFS(os.DirFS(out))
		reload := make(chan struct{})
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
		err = watcher.Add(src)
		if err != nil {
			log.Fatal(err)
		}
		err = watcher.Add(filepath.Join(src, "blog/"))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(watcher.WatchList())
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.URL.Path)
			switch r.URL.Path {
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
						err = build(src, out)
						if err != nil {
							println(err.Error())
						}
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

			default:
				fsv.ServeHTTP(w, r)
			}
		})
		println("http://localhost:8081")
		err = http.ListenAndServe(":8081", h)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err := build(src, out)
	if err != nil {
		fail(err)
	}
}

func build(src, dst string) error {
	os.MkdirAll(dst, 0755)
	err := writeDocs(src, dst)
	if err != nil {
		return err
	}
	return writeBlog(src, dst)
}
