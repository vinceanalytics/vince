package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"
)

var files atomic.Int64
var dirs atomic.Int64

func main() {
	flag.Parse()
	r := &entry{
		path: flag.Arg(0),
	}
	e := empty()
	if !filepath.IsAbs(r.path) {
		var b string
		for _, d := range filepath.SplitList(r.path) {
			b = filepath.Join(b, d)
			if o, err := os.ReadFile(filepath.Join(b, ".gitignore")); err == nil {
				e = e.merge(parse(bytes.NewReader(o)))
			}
		}
	}
	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()

	walk(&wg, r, e, flag.Arg(0))
	wg.Wait()
	elapsed := time.Since(start)
	summary(r, elapsed)
}

var _ sort.Interface = (*list)(nil)

type list []*entry

func (ls list) Len() int {
	return len(ls)
}

func (ls list) Less(i, j int) bool {
	return ls[i].size < ls[j].size
}

func (ls list) Swap(i, j int) {
	ls[i], ls[j] = ls[j], ls[i]
}

type entry struct {
	parent   *entry
	path     string
	mu       sync.Mutex
	size     int64
	children []*entry
}

func (e *entry) walk(f func(*entry)) {
	f(e)
	for _, v := range e.children {
		v.walk(f)
	}
}

func (e *entry) update(size int64) {
	e.mu.Lock()
	e.size += size
	e.mu.Unlock()
	if e.parent != nil {
		e.parent.update(size)
	}
}
func (e *entry) add(ch *entry) {
	e.mu.Lock()
	e.children = append(e.children, ch)
	e.mu.Unlock()
}

func walk(wg *sync.WaitGroup, parent *entry, r *rules, path string) {
	dirs.Add(1)
	defer wg.Done()
	if b, err := os.ReadFile(filepath.Join(path, ".gitignore")); err == nil {
		r = r.merge(parse(bytes.NewReader(b)))
	}
	e, err := os.ReadDir(path)
	if err != nil {
		println("> ERROR", path, err.Error())
		return
	}
	x := &entry{
		parent: parent,
		path:   path,
	}
	parent.add(x)
	for _, f := range e {
		i, err := f.Info()
		if err != nil {
			println("> ERROR", path, err.Error())
			continue
		}
		name := f.Name()
		if r.skip(name, i) {
			continue
		}
		if i.IsDir() {
			wg.Add(1)
			go walk(wg, x, r, filepath.Join(path, f.Name()))
		} else {
			files.Add(1)
			x.update(i.Size())
		}
	}
}

const (
	kb = 1 << 10
	mb = 1 << 20
	gb = 1 << 30
)

func human(x int64) string {
	switch {
	case x < mb:
		return fmt.Sprintf("%dkb", x/kb)
	case x < gb:
		return fmt.Sprintf("%dmb", x/mb)
	default:
		return fmt.Sprintf("%dgb", x/gb)
	}
}

func count(x int64) string {
	switch {
	case x < 1000:
		return fmt.Sprintf("%d", x)
	default:
		return fmt.Sprintf("%dk", x/1000)
	}
}

func summary(e *entry, elapsed time.Duration) {
	ls := make(list, 0, 3<<10)
	e.walk(func(e *entry) {
		ls = append(ls, e)
	})
	ls = ls[1:]
	sort.Sort(sort.Reverse(ls))
	if len(ls) > 10 {
		ls = ls[:10]
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight)

	fmt.Fprintf(w, "scanned\t\n")
	fmt.Fprintf(w, " files\t%s\t\n", count(files.Load()))
	fmt.Fprintf(w, " dirs\t%s\t\n", count(dirs.Load()))
	fmt.Fprintf(w, " in \t%s\t\n", elapsed)
	fmt.Fprintln(w, " \n---\t---\t")

	for _, v := range ls {
		fmt.Fprintf(w, "%s\t%s\t\n", v.path, human(v.size))
	}
	w.Flush()
}

type rules struct {
	patterns []*pattern
}

func parse(b io.Reader) *rules {
	e := empty()
	s := bufio.NewScanner(b)
	for s.Scan() {
		e.parse(s.Text())
	}
	return e
}

func empty() *rules {
	e := &rules{}
	e.parse(".git")
	return e
}

func (r *rules) merge(o *rules) *rules {
	return &rules{patterns: append(r.patterns, o.patterns...)}
}

func (r *rules) skip(path string, fi os.FileInfo) bool {
	// Don't match on empty dirs.
	if path == "" {
		return false
	}

	if path == "." || path == "./" {
		return false
	}
	for _, p := range r.patterns {
		if p.match == nil {
			log.Printf("ignore: no matcher supplied for %q", p.raw)
			return false
		}
		if p.negate {
			if p.mustDir && !fi.IsDir() {
				return true
			}
			if !p.match(path, fi) {
				return true
			}
			continue
		}
		if p.mustDir && !fi.IsDir() {
			continue
		}
		if p.match(path, fi) {
			return true
		}
	}
	return false
}

func (r *rules) parse(rl string) error {
	rl = strings.TrimSpace(rl)
	if rl == "" {
		return nil
	}
	if strings.HasPrefix(rl, "#") {
		return nil
	}
	if strings.Contains(rl, "**") {
		return errors.New("double-star (**) syntax is not supported")
	}
	if _, err := filepath.Match(rl, "abc"); err != nil {
		return err
	}

	p := &pattern{raw: rl}
	if strings.HasPrefix(rl, "!") {
		p.negate = true
		rl = rl[1:]
	}
	if strings.HasSuffix(rl, "/") {
		p.mustDir = true
		rl = strings.TrimSuffix(rl, "/")
	}

	if strings.HasPrefix(rl, "/") {
		p.match = func(n string, fi os.FileInfo) bool {
			rl = strings.TrimPrefix(rl, "/")
			ok, err := filepath.Match(rl, n)
			if err != nil {
				log.Printf("Failed to compile %q: %s", rl, err)
				return false
			}
			return ok
		}
	} else if strings.Contains(rl, "/") {
		p.match = func(n string, fi os.FileInfo) bool {
			ok, err := filepath.Match(rl, n)
			if err != nil {
				log.Printf("Failed to compile %q: %s", rl, err)
				return false
			}
			return ok
		}
	} else {
		p.match = func(n string, fi os.FileInfo) bool {
			n = filepath.Base(n)
			ok, err := filepath.Match(rl, n)
			if err != nil {
				log.Printf("Failed to compile %q: %s", rl, err)
				return false
			}
			return ok
		}
	}
	r.patterns = append(r.patterns, p)
	return nil
}

type matcher func(name string, fi os.FileInfo) bool

type pattern struct {
	raw     string
	match   matcher
	negate  bool
	mustDir bool
}
