package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	artifacts     = "dist/artifacts.json"
	metadata      = "dist/metadata.json"
	labelCreated  = "org.opencontainers.image.created=%q"
	labelRevision = "org.opencontainers.image.revision=%q"
	labelTitle    = "org.opencontainers.image.title=%q"
)

func main() {
	var a []Artifact
	read(artifacts, &a)
	var meta MetaData
	read(metadata, &meta)
	dockerFile, err := os.ReadFile("Containerfile")
	if err != nil {
		log.Fatal(err)
	}
	// var amend []string
	for _, v := range a {
		if v.Type != "Binary" || v.Os != "linux" {
			continue
		}
		// copy docker file to the context
		os.WriteFile(filepath.Join(filepath.Dir(v.Path), "Containerfile"), dockerFile, 0600)
		cmd := exec.Command(
			"podman", "build",
			"--log-level", "debug",
			"--platform", v.Os+"/"+v.Arch,
			"--label", fmt.Sprintf(labelCreated, meta.Date.UTC().Format(time.RFC3339)),
			"--label", fmt.Sprintf(labelRevision, meta.Commit),
			"--label", fmt.Sprintf(labelTitle, meta.Image(v.Arch)),
			"--manifest", meta.Manifest(),
			filepath.Dir(v.Path),
		)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
	cmd := exec.Command("podman", "tag", meta.Manifest(), meta.Latest())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
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
	return fmt.Sprintf("vinceanalytics/%s:%s-%s", m.Name, m.Tag, arch)
}

func (m MetaData) Manifest() string {
	return fmt.Sprintf("vinceanalytics/%s:%s", m.Name, m.Tag)
}

func (m MetaData) Latest() string {
	return fmt.Sprintf("vinceanalytics/%s:latest", m.Name)
}

func read(path string, o any) {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(b, o)
	if err != nil {
		log.Fatal(err)
	}
}
