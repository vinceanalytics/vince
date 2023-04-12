package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	flag.Parse()
	src := flag.Arg(0)
	if src == "" {
		return
	}
	_, err := os.Stat(src)
	if os.IsNotExist(err) {
		return
	}
	dest := flag.Arg(1)
	if dest == "" {
		return
	}
	dest, err = filepath.Abs(dest)
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			path, _ = filepath.Rel(src, path)
			if path != "" {
				os.Mkdir(filepath.Join(dest, path), info.Mode())
			}
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		path, _ = filepath.Rel(src, path)
		return os.WriteFile(filepath.Join(dest, path), b, info.Mode())
	})
	if err != nil {
		log.Fatal(err)
	}
}
