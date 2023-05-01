package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	_ "embed"

	"github.com/gernest/vince/tools"
)

var root string

//go:embed README.tmpl
var readmeBytes []byte

var tpl = template.Must(template.New("root").
	Funcs(template.FuncMap{
		"releaseTable": releaseTable,
	}).
	Parse(string(readmeBytes)))

var project tools.Project

func main() {
	println("### Generating README.md with release info ###")
	var err error
	root, err = filepath.Abs("../")
	if err != nil {
		tools.Exit("failed to resolve root ", err.Error())
	}
	project = tools.Release(root)
	println(">>> building man pages")
	man := tools.ExecCollect(
		"go", "run", filepath.Join(root, "main.go"), "man",
	)
	tools.WriteFile(filepath.Join(root, "man", "vince.1"), []byte(man))
	man = tools.ExecCollect(
		"go", "run", filepath.Join(root, "cmd", "v8s", "main.go"), "man",
	)
	tools.WriteFile(filepath.Join(root, "man", "v8s.1"), []byte(man))
	make()
}

func make() {
	var o bytes.Buffer
	err := tpl.Execute(&o, map[string]any{
		"Project": &project,
	})
	if err != nil {
		tools.Exit("failed to render release readme", err.Error())
	}
	tools.WriteFile(filepath.Join(root, "README.md"), o.Bytes())
}

func releaseTable(artifacts []tools.Artifact) string {
	var table tools.Table
	table.Init(
		"filename", "signature", "size",
	)
	for _, a := range artifacts {
		if a.Type != "Archive" {
			continue
		}
		stat, err := os.Stat(filepath.Join(root, a.Path))
		if err != nil {
			tools.Exit("can't find artifact", err.Error())
		}
		table.Row(
			fmt.Sprintf("[%s](%s)", a.Name, link(project.Meta.Tag, a.Name)),
			fmt.Sprintf("[minisig](%s)", link(project.Meta.Tag, a.Name+".minisig")),
			fmt.Sprintf("`%s`", size(int(stat.Size()))),
		)
	}
	table.Flush()
	return table.String()
}

func link(tag, name string) string {
	return fmt.Sprintf("https://github.com/vinceanalytics/vince/releases/download/%s/%s", tag, name)
}
func size(n int) string {
	if n < (1 << 20) {
		return strconv.Itoa(n/(1<<10)) + "kb"
	}
	if n < (1 << 30) {
		return strconv.Itoa(n/(1<<20)) + "mb"
	}
	return strconv.Itoa(n)
}
