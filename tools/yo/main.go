package main

import (
	"flag"
	"fmt"
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
	var wg sync.WaitGroup
	wg.Add(1)
	start := time.Now()

	walk(&wg, r, flag.Arg(0))
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

func walk(wg *sync.WaitGroup, parent *entry, path string) {
	dirs.Add(1)
	defer wg.Done()
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
		if strings.HasPrefix(f.Name(), ".") {
			return
		}
		if i.IsDir() {
			wg.Add(1)
			go walk(wg, x, filepath.Join(path, f.Name()))
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
