package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/tools"
	"github.com/vinceanalytics/vince/v8s/cmd/v8s"
	"moul.io/http2curl"
)

func main() {
	if os.Getenv("DOCS") != "" {
		root := tools.RootVince()
		mannPage(root, vince.App())
		mannPage(root, v8s.App())
		completion(root)
		guides(root)
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

func guides(root string) {
	guideCreateSite(root)
}

const (
	host   = "http://localhost:8080"
	bearer = "$VINCE_BOOTSTRAP_KEY"
	domain = "vince.example.com"
)

func guideCreateSite(root string) {
	b, _ := json.Marshal(map[string]string{
		"domain": domain,
	})
	r, _ := http.NewRequest(http.MethodPost, host+"/api/v1/sites", bytes.NewReader(b))
	r.Header.Set("Authorization", "Bearer "+bearer)

	write(filepath.Join(root, "website/docs/guide/files/create_site.sh"), r)
}

func write(path string, r *http.Request) {
	cmd, _ := http2curl.GetCurlCommand(r)
	tools.WriteFile(path, []byte(breakDown(cmd.String())))
}

func breakDown(a string) string {
	a = strings.ReplaceAll(a, "' -d '", "' \\\n\t-d '")
	a = strings.ReplaceAll(a, "' -H '", "' \\\n\t-H '")
	return a
}
