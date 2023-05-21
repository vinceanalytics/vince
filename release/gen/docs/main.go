package main

import (
	"fmt"
	"os"
	"path/filepath"

	_ "embed"

	"github.com/gernest/vince/cmd/app/v8s"
	"github.com/gernest/vince/cmd/app/vince"
	"github.com/gernest/vince/tools"
	"github.com/urfave/cli/v3"
)

func main() {
	root := tools.RootVince()
	mannPage(root, vince.App())
	mannPage(root, v8s.App())
	cliPage(root, vince.App())
	cliPage(root, v8s.App())
	completion(root)
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

func mannPage(root string, app *cli.App) {
	println("> man page", app.Name)
	m, err := app.ToMan()
	if err != nil {
		tools.Exit(err.Error())
	}
	tools.WriteFile(filepath.Join(root, "man", app.Name+".1"), []byte(m))
}

func cliPage(root string, app *cli.App) {
	println("> cli ", app.Name)
	m, err := app.ToMarkdown()
	if err != nil {
		tools.Exit(err.Error())
	}
	tools.WriteFile(filepath.Join(root, "docs", "guide", fmt.Sprintf("cli-%s.md", app.Name)), []byte(m))
}
