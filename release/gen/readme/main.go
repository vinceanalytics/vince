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

var tpl = template.Must(template.New("root").Parse(string(readmeBytes)))

var meta tools.MetaData
var artifacts []tools.Artifact

func main() {
	println("### Generating README.md with release info ###")
	var err error
	root, err = filepath.Abs("../")
	if err != nil {
		tools.Exit("failed to resolve root ", err.Error())
	}
	meta, artifacts = tools.Release(root)
	make()
}

func make() {
	var o bytes.Buffer
	err := tpl.Execute(&o, map[string]any{
		"ReleaseTable": releaseTable(),
		"Release":      meta,
	})
	if err != nil {
		tools.Exit("failed to render release readme", err.Error())
	}
	tools.WriteFile(filepath.Join(root, "README.md"), o.Bytes())
}

func releaseTable() string {
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
			fmt.Sprintf("[%s](%s)", a.Name, meta.Link(a.Name)),
			fmt.Sprintf("[minisig](%s)", meta.Link(a.Name+".minisig")),
			fmt.Sprintf("`%s`", size(int(stat.Size()))),
		)
	}
	table.Flush()
	return table.String()
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

type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
}
