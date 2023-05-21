package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gernest/vince/tools"
	"golang.org/x/mod/semver"
)

func main() {
	flag.Parse()
	v := latestTag()
	if v == "" {
		v = "v0.0.0"
	}
	if !semver.IsValid(v) {
		tools.Exit("VERSION must be in vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]] format")
	}
	parts := breakDown(v)
	switch os.Getenv("VERSION") {
	case "major":
		i, err := strconv.Atoi(strings.TrimPrefix(parts[0], "v"))
		if err != nil {
			tools.Exit("  failed parsing major version", parts[0], err.Error())
		}
		i++
		parts[0] = "v" + strconv.Itoa(i)
	case "minor":
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			tools.Exit("  failed parsing minor version", parts[1], err.Error())
		}
		i++
		parts[1] = strconv.Itoa(i)
	case "patch":
		i, err := strconv.Atoi(parts[2])
		if err != nil {
			tools.Exit("  failed parsing patch version", parts[2], err.Error())
		}
		i++
		parts[2] = strconv.Itoa(i)

	}
	pre := os.Getenv("PRERELEASE")
	if pre != "" {
		if len(parts) == 3 {
			parts = append(parts, pre)
		} else {
			parts[3] = pre
		}
	}
	v = format(parts)
	build(v)
	commit(v)
	tag(v)
}

func format(p []string) string {
	s := strings.Join(p[:3], ".")
	if len(p) == 4 {
		s += "-" + p[3]
	}
	return s
}

func breakDown(v string) (o []string) {
	if !semver.IsValid(v) {
		tools.Exit("VERSION must be in vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]] format")
	}
	a := semver.MajorMinor(v)
	o = strings.Split(a, ".")
	patch, rest, found := strings.Cut(strings.TrimPrefix(v, a), ".")
	if found && patch != "" {
		o = append(o, patch, rest)
	} else {
		o = append(o, rest)
	}
	return
}

func build(v string) {
	root := tools.RootVince()
	println("> using", v)
	chart := tools.ReadFile(filepath.Join(root, "chart/Chart.yaml"))
	var o bytes.Buffer
	s := bufio.NewScanner(bytes.NewReader(chart))
	chartVersion := "version: "
	appVersion := "appVersion: "
	for s.Scan() {
		text := s.Text()
		if strings.HasPrefix(text, chartVersion) {
			text = strings.TrimPrefix(text, chartVersion)
			text = strings.TrimSpace(text)
			text = strings.TrimPrefix(text, "\"")
			text = strings.TrimSuffix(text, "\"")
			if text == v {
				println("> no version changes")
				return
			}
			switch semver.Compare(v, text) {
			case 0:
				println("> no version changes")
				return
			case -1:
				tools.Exit("VERSION must be greater than", text)
			case 1:
				text = chartVersion + v
			}

		}
		if strings.HasPrefix(text, appVersion) {
			text = appVersion + v
		}
		fmt.Fprintln(&o, text)
	}
	tools.WriteFile(filepath.Join(root, "chart/Chart.yaml"), o.Bytes())
	app := tools.ReadFile(filepath.Join(root, "pkg/version/VERSION.txt"))
	switch semver.Compare(v, string(app)) {
	case 0:
		println("> no version changes")
		return
	case -1:
		tools.Exit("VERSION must be greater than", string(app))
	}
	tools.WriteFile(filepath.Join(root, "pkg/version/VERSION.txt"), []byte(v))
}

func commit(v string) {
	tools.ExecPlain(
		"git", "commit", "-am", "release "+v,
	)
}

func tag(v string) {
	tools.ExecPlain(
		"git", "tag", "-a", v, "-m", "release "+v,
	)
}

func latestTag() string {
	return tools.ExecCollect("git", "describe", "--abbrev=0")
}
