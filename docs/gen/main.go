package main

import (
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func main() {
	flag.Parse()
	src := flag.Arg(0)
	if src == "" {
		return
	}
	var stats []Stat
	filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		path, _ = filepath.Rel(src, path)
		stats = append(stats, Stat{
			Path: path,
			Mod:  info.ModTime(),
		})
		return nil
	})
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Path < stats[j].Path
	})
	b, _ := json.Marshal(stats)
	os.WriteFile(filepath.Join(src, "manifest.json"), b, 0600)
}

type Stat struct {
	Path string
	Mod  time.Time
}
