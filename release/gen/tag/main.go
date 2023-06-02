package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vinceanalytics/vince/tools"
	"golang.org/x/mod/semver"
)

func main() {
	flag.Parse()
	v := tools.Version()
	build(v)
	commit(v)
	tag(v)
}

func build(v string) {
	root := tools.RootVince()
	println("> using", v)
	updateVersionPackage(root, v)
	updateHelmCharts(root, v)
}

func updateVersionPackage(root, v string) {
	println("> update vince version")
	app := tools.ReadFile(filepath.Join(root, "pkg/version/VERSION.txt"))
	switch semver.Compare(v, string(app)) {
	case 0:
		println("> no version changes")
		os.Exit(0)
	case -1:
		tools.Exit(v, "VERSION must be greater than", string(app))
	}
	tools.WriteFile(filepath.Join(root, "pkg/version/VERSION.txt"), []byte(v))
}

func updateHelmCharts(root, v string) {
	println("> update helm charts")
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
}

func commit(v string) {
	println("> commit", v)
	tools.ExecPlain(
		"git", "commit", "-am", "release "+v,
	)
}

func tag(v string) {
	println("> tag", v)
	tools.ExecPlain(
		"git", "tag", "-a", v, "-m", "release "+v,
	)
	tools.ExecPlain(
		"git", "push",
	)
}
