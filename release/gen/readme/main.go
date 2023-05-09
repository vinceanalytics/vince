package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	_ "embed"

	"github.com/gernest/vince/cmd/app/v8s"
	"github.com/gernest/vince/cmd/app/vince"
	"github.com/gernest/vince/tools"
	"github.com/urfave/cli/v3"
)

var root string

//go:embed install.tmpl
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
	readme()

	println(">>> building  pages")
	mannPage(vince.App())
	mannPage(v8s.App())
	cliPage(vince.App())
	cliPage(v8s.App())
	completion()
}

func readme() {
	var o bytes.Buffer
	err := tpl.Execute(&o, map[string]any{
		"Project": &project,
	})
	if err != nil {
		tools.Exit("failed to render release readme", err.Error())
	}
	tools.WriteFile(filepath.Join(root, "docs", "guide", "install.md"), o.Bytes())
}

func releaseTable(artifacts []tools.Artifact) string {
	var table tools.Table
	table.Init(
		"filename", "signature", "size",
	)
	for _, a := range artifacts {
		switch a.Type {
		case "Linux Package", "Archive":
		default:
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

func completion() {
	println("### Generating autocomplete scripts ###")
	base := tools.ModuleRoot("github.com/urfave/cli/v3")
	println(">>> from ", base)
	bashFile := filepath.Join(base, "autocomplete", "bash_autocomplete")
	bash, err := os.ReadFile(bashFile)
	if err != nil {
		tools.Exit(err.Error())
	}
	powerFile := filepath.Join(base, "autocomplete", "powershell_autocomplete.ps1")
	power, err := os.ReadFile(powerFile)
	if err != nil {
		tools.Exit(err.Error())
	}
	zshFile := filepath.Join(base, "autocomplete", "zsh_autocomplete")
	zsh, err := os.ReadFile(zshFile)
	if err != nil {
		tools.Exit(err.Error())
	}
	binaries := []string{"vince", "v8s"}
	for _, name := range binaries {
		fileBash := filepath.Join(root, "completions", name, name+".bash")
		fileZsh := filepath.Join(root, "completions", name, name+".zsh")
		filePowerShell := filepath.Join(root, "completions", name, name+".ps1")
		os.MkdirAll(filepath.Join(root, "completions", name), 0700)
		tools.WriteFile(fileBash, bash)
		tools.WriteFile(filePowerShell, power)
		tools.WriteFile(fileZsh, zsh)
	}
	vinceFish, err := vince.App().ToFishCompletion()
	if err != nil {
		tools.Exit(err.Error())
	}
	v8sFish, err := v8s.App().ToFishCompletion()
	if err != nil {
		tools.Exit(err.Error())
	}
	tools.WriteFile(filepath.Join(root, "completions", "vince", "vince.fish"), []byte(vinceFish))
	tools.WriteFile(filepath.Join(root, "completions", "v8s", "v8s.fish"), []byte(v8sFish))
}

func mannPage(app *cli.App) {
	m, err := app.ToMan()
	if err != nil {
		tools.Exit(err.Error())
	}
	tools.WriteFile(filepath.Join(root, "man", app.Name+".1"), []byte(m))
}

func cliPage(app *cli.App) {
	m, err := app.ToMarkdown()
	if err != nil {
		tools.Exit(err.Error())
	}
	tools.WriteFile(filepath.Join(root, "docs", "guide", app.Name+"-cli.md"), []byte(m))
}
