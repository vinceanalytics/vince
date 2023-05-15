package tools

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"gopkg.in/yaml.v2"
)

func ExecCollect(name string, args ...string) string {
	var o bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &o
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(o.String())
}

func ExecPlain(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func ExecPlainWithWorkingPath(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// ReadUA regex file by name from https://github.com/matomo-org/device-detector
// root directory.
//
// Env var DEVICE_DETECTOR_ROOT is used to set the root path.
func ReadUA(name string, out any) {
	path := filepath.Join(os.Getenv("DEVICE_DETECTOR_ROOT"), "regexes", name)
	f, err := os.ReadFile(path)
	if err != nil {
		println(">>> failed to read  ", path, err.Error())
		println("    set DEVICE_DETECTOR_ROOT")
		println("    to path where you cloned ", "https://github.com/matomo-org/device-detector")
		os.Exit(1)
	}
	err = yaml.Unmarshal(f, out)
	if err != nil {
		Exit("failed to  decode ", path, err.Error())
	}
}

func WriteFile(path string, data []byte) {
	err := os.WriteFile(path, data, 0600)
	if err != nil {
		Exit("failed to write ", path, err.Error())
	}
	println("    write: ", path)
}

func Remove(path string) error {
	err := os.Remove(path)
	if err != nil {
		Exit("failed to delete file ", path, err.Error())
	}
	println("    delete: ", path)
	return nil
}

func Exit(a ...string) {
	println(">>> ", strings.Join(a, " "))
	os.Exit(1)
}

// Table generates nice looking markdown tables
type Table struct {
	bytes.Buffer
	b    bytes.Buffer
	line bytes.Buffer
	tw   tabwriter.Writer
}

func (t *Table) Init(header ...string) *Table {
	t.b.Reset()
	t.tw.Init(&t.b, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	dash := make([]string, len(header))
	for i := range dash {
		dash[i] = "----"
	}
	t.Row(header...)
	t.Row(dash...)
	return t
}

func (t *Table) Row(entries ...string) {
	t.line.Reset()
	for i := range entries {
		if i != 0 {
			t.line.WriteRune('\t')
		}
		t.line.WriteString(entries[i])
	}
	t.line.WriteByte('\n')
	io.Copy(&t.tw, &t.line)
}

func (t *Table) Flush() {
	t.tw.Flush()
	t.Buffer.Reset()
	s := bufio.NewScanner(&t.b)
	for s.Scan() {
		fmt.Fprintf(&t.Buffer, "|%s|\n", s.Text())
	}
}

type Artifact struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Path  string `json:"path"`
	Os    string `json:"goos"`
	Arch  string `json:"goarch"`
	Extra struct {
		ID           string `json:"id"`
		DockerConfig struct {
			ID string `json:"id"`
		} `json:"DockerConfig"`
	} `json:"extra"`
}

type MetaData struct {
	Name    string    `json:"project_name"`
	Tag     string    `json:"tag"`
	Version string    `json:"version"`
	Commit  string    `json:"commit"`
	Date    time.Time `json:"date"`
}

type Project struct {
	Meta      MetaData
	Artifacts []*Artifacts
}

type Artifacts struct {
	ID        string
	Artifacts []Artifact
}

func Release(root string) (p Project) {
	readJSON(filepath.Join(root, "dist/metadata.json"), &p.Meta)
	var artifacts []Artifact
	readJSON(filepath.Join(root, "dist/artifacts.json"), &artifacts)
	m := make(map[string][]Artifact)
	for _, a := range artifacts {
		switch a.Type {
		case "Archive", "Linux Package":
			m[a.Extra.ID] = append(m[a.Extra.ID], a)
		}
	}
	for k, v := range m {
		p.Artifacts = append(p.Artifacts, &Artifacts{
			ID:        k,
			Artifacts: v,
		})
	}
	sort.Slice(p.Artifacts, func(i, j int) bool {
		// This is to ensure vince comes before v8s
		return p.Artifacts[i].ID > p.Artifacts[j].ID
	})
	return
}

func readJSON(path string, o any) {
	b, err := os.ReadFile(path)
	if err != nil {
		Exit("failed to read", path, err.Error())
	}
	err = json.Unmarshal(b, o)
	if err != nil {
		Exit("failed to decode", path, err.Error())
	}
}

func Package(module string) *debug.Module {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		Exit("failed to read build info")
	}
	for _, m := range build.Deps {
		if m.Path == module {
			return m
		}
	}
	Exit("no such module ", module)
	return nil
}

func RootVince() string {
	return Root("github.com/gernest/vince")
}

func Root(pkg string) string {
	return ExecCollect("go", "list", "-m", "-f", "{{.Dir}}", pkg)
}

func Pkg(pkg string) string {
	return ExecCollect("go", "list", "-m", pkg)
}
