package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gernest/vince/tools"
	"golang.org/x/mod/semver"
)

func main() {
	root := tools.RootVince()
	v := os.Getenv("VERSION")
	if v == "" {
		tools.Exit("VERSION env must be set")
	}
	if !semver.IsValid(v) {
		tools.Exit("VERSION must be in vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]] format")
	}
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
}
