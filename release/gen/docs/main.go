package main

import (
	"os"
	"path/filepath"

	_ "embed"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/internal/tools"
)

func main() {
	if os.Getenv("DOCS") != "" {
		root := tools.RootVince()
		mannPage(root, vince.App())
		completion(root)
	}
}

func completion(root string) {
	println("> completions")
	base := tools.Root("github.com/urfave/cli/v3")
	bashFile := filepath.Join(base, "autocomplete", "bash_autocomplete")
	bash := tools.ReadFile(bashFile)
	powerFile := filepath.Join(base, "autocomplete", "powershell_autocomplete.ps1")
	power := tools.ReadFile(powerFile)
	zshFile := filepath.Join(base, "autocomplete", "zsh_autocomplete")
	zsh := tools.ReadFile(zshFile)
	binaries := []string{"vince"}
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
	tools.WriteFile(filepath.Join(root, "completions", "vince", "vince.fish"), []byte(vinceFish))
}

func mannPage(root string, app *cli.Command) {
	println("> man page", app.Name)
	m, err := app.ToMan()
	if err != nil {
		tools.Exit(err.Error())
	}
	tools.WriteFile(filepath.Join(root, "man", app.Name+".1"), []byte(m))
}
