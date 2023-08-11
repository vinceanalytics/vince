package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/urfave/cli/v3"
	"github.com/vinceanalytics/vince/internal/cmd/vince"
	"github.com/vinceanalytics/vince/tools"
	"github.com/vinceanalytics/vince/v8s/cmd/v8s"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func main() {
	if os.Getenv("DOCS") != "" {
		root := tools.RootVince()
		mannPage(root, vince.App())
		mannPage(root, v8s.App())
		completion(root)
		config(vince.App())
	}
}

func config(a *cli.App) {
	var b bytes.Buffer
	b.WriteString(`---
title: Configuration
---

# Configuration

**Vince** acn be configured with commandline flags or environment variables.

::: tip
We recommend the use of environment variables. They are safer (passing secrets ),
and it allows simpler deployments, where you set the environment and just call 
**vince** command.
:::

`)

	a.Setup()
	fmt.Fprintln(&b)
	var buf bytes.Buffer
	title := cases.Title(language.English)
	for _, c := range a.VisibleFlagCategories() {
		fmt.Fprintf(&b, "# %s\n\n", title.String(c.Name()))
		set := flag.NewFlagSet("", flag.ContinueOnError)
		m := make(map[string]cli.VisibleFlag)
		for _, f := range c.Flags() {
			f.Apply(set)
			m[f.Names()[0]] = f
		}
		set.VisitAll(func(f *flag.Flag) {
			o, ok := m[f.Name]
			if !ok {
				return
			}
			value := f.Value.String()
			if value == "" {
				value = "..."
			}
			switch e := o.(type) {
			case *cli.StringFlag:
				renderFlag(&b, f.Name, e.Usage, value, e.EnvVars[0])
			case *cli.DurationFlag:
				renderFlag(&b, f.Name, e.Usage, value, e.EnvVars[0])
			case *cli.BoolFlag:
				renderFlag(&b, f.Name, e.Usage, value, e.EnvVars[0])
			case *cli.IntFlag:
				renderFlag(&b, f.Name, e.Usage, value, e.EnvVars[0])
			case *cli.StringSliceFlag:
				value = strings.Join(e.Value, ",")
				renderFlag(&b, f.Name, e.Usage, value, e.EnvVars[0])
			case *cli.Uint64SliceFlag:
				renderFlag(&b, f.Name, e.Usage, list(&buf, e.Value), e.EnvVars[0])
			}
		})
	}
	tools.WriteFile(
		filepath.Join(tools.RootVince(), "website/docs/guide/config.md"),
		b.Bytes(),
	)
}

func list[T any](o *bytes.Buffer, ls []T) string {
	if len(ls) == 0 {
		return "..."
	}
	o.Reset()
	for i, v := range ls {
		if i != 0 {
			o.WriteByte('\n')
		}
		fmt.Fprint(o, v)
	}
	return o.String()
}

func renderFlag(o io.Writer, name, usage, value, env string) {
	fmt.Fprintf(o, "## %s\n%s\n", name, usage)
	value = scrub(name, value)
	fmt.Fprintln(o, "::: code-group")
	fmt.Fprintf(o, "```shell [flag]\n--%s=%q\n```\n", name, value)
	fmt.Fprintf(o, "```shell [env]\n%s=%q\n```\n", env, value)
	fmt.Fprintln(o, ":::")
}

func scrub(name, value string) string {
	switch name {
	case "bootstrap-email",
		"bootstrap-key",
		"bootstrap-password",
		"bootstrap-name",
		"secret", "secret-age":
		return "..."
	default:
		return value
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
