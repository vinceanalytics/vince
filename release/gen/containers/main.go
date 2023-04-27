package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gernest/vince/tools"
)

const (
	artifacts     = "dist/artifacts.json"
	metadata      = "dist/metadata.json"
	labelCreated  = "org.opencontainers.image.created=%q"
	labelRevision = "org.opencontainers.image.revision=%q"
	labelTitle    = "org.opencontainers.image.title=%q"
)

var root string

func main() {
	println("### Building container image ###")
	var err error
	root, err = filepath.Abs("../")
	if err != nil {
		tools.Exit("failed to resolve project root", err.Error())
	}
	println(">>> root: ", root)
	make()
}

func make() {
	var a []Artifact
	read(filepath.Join(root, artifacts), &a)
	var meta MetaData
	read(filepath.Join(root, metadata), &meta)
	dockerFile, err := os.ReadFile(filepath.Join(root, "Dockerfile"))
	if err != nil {
		tools.Exit("failed to read docker  file", err.Error())
	}
	for _, v := range a {
		if v.Type != "Binary" || v.Os != "linux" {
			continue
		}
		// copy docker file to the context
		tools.WriteFile(filepath.Join(filepath.Dir(v.Path), "Dockerfile"), dockerFile)

		tools.ExecPlain(
			"docker",
			"--log-level", "info",
			"build",
			"--platform", v.Os+"/"+v.Arch,
			"--label", fmt.Sprintf(labelCreated, meta.Date.UTC().Format(time.RFC3339)),
			"--label", fmt.Sprintf(labelRevision, meta.Commit),
			"--label", fmt.Sprintf(labelTitle, meta.Image(v.Arch)),
			"--tag", meta.Image(v.Arch),
			filepath.Dir(v.Path),
		)
	}
	tools.ExecPlain("docker", "tag", meta.Image(runtime.GOARCH), meta.Latest())
}

type Artifact struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Path string `json:"path"`
	Os   string `json:"goos"`
	Arch string `json:"goarch"`
}

type MetaData struct {
	Name    string    `json:"project_name"`
	Tag     string    `json:"tag"`
	Version string    `json:"version"`
	Commit  string    `json:"commit"`
	Date    time.Time `json:"date"`
}

func (m MetaData) Image(arch string) string {
	return fmt.Sprintf("ghcr.io/vinceanalytics/%s:%s-%s", m.Name, m.Tag, arch)
}

func (m MetaData) Manifest() string {
	return fmt.Sprintf("ghcr.io/vinceanalytics/%s:%s", m.Name, m.Tag)
}

func (m MetaData) Latest() string {
	return fmt.Sprintf("ghcr.io/vinceanalytics/%s:latest", m.Name)
}

func read(path string, o any) {
	b, err := os.ReadFile(path)
	if err != nil {
		tools.Exit("failed to read", path, err.Error())
	}
	err = json.Unmarshal(b, o)
	if err != nil {
		tools.Exit("failed to decode", path, err.Error())
	}
}
