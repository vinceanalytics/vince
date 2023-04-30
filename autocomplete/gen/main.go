package main

import (
	"os"
	"path/filepath"

	"github.com/gernest/vince/tools"
	_ "github.com/urfave/cli/v3"
)

func main() {
	println("### Generating autocomplete scripts ###")
	root := tools.ModuleRoot("github.com/urfave/cli/v3")
	println(">>> from ", root)
	bashFile := filepath.Join(root, "autocomplete", "bash_autocomplete")
	bash, err := os.ReadFile(bashFile)
	if err != nil {
		tools.Exit(err.Error())
	}
	powerFile := filepath.Join(root, "autocomplete", "powershell_autocomplete.ps1")
	power, err := os.ReadFile(powerFile)
	if err != nil {
		tools.Exit(err.Error())
	}
	zshFile := filepath.Join(root, "autocomplete", "zsh_autocomplete")
	zsh, err := os.ReadFile(zshFile)
	if err != nil {
		tools.Exit(err.Error())
	}
	os.MkdirAll("completions/bash", 0700)
	os.MkdirAll("completions/powershell", 0700)
	os.MkdirAll("completions/fish", 0700)
	binaries := []string{"vince", "v8s"}
	for _, name := range binaries {
		tools.WriteFile(filepath.Join("completions/bash", name), bash)
		tools.WriteFile(filepath.Join("completions/powershell", name+".ps1"), power)
		tools.WriteFile(filepath.Join("completions/fish", name), zsh)
	}
}
