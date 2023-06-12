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
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

func ExecCollect(name string, args ...string) string {
	var o bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
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
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func ExecPlainWithWorkingPath(dir, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = dir
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func ExecPlainWith(f func(*exec.Cmd), name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	f(cmd)
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

func MkDir(path string) {
	ExecPlain("mkdir", "-pv", path)
}

func ReadFile(path string) []byte {
	println("    read: ", path)
	b, err := os.ReadFile(path)
	if err != nil {
		Exit("failed to read ", path, err.Error())
	}
	return b
}

func Remove(path string) {
	println("    delete: ", path)
	ExecPlain("rm", "-f", path)
}

func CopyFile(src, dest string) {
	ExecPlain("cp", "-v", src, dest)
}

func CopyDir(src, dest string, workingDir ...string) {
	if len(workingDir) > 0 {
		ExecPlainWithWorkingPath(workingDir[0], "cp", "-rv", src, dest)
	} else {
		ExecPlain("cp", "-rv", src, dest)
	}
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

var root string
var once sync.Once

func RootVince() string {
	once.Do(func() {
		root = Root("github.com/vinceanalytics/vince")
	})
	return root
}

func Root(pkg string) string {
	return ExecCollect("go", "list", "-m", "-f", "{{.Dir}}", pkg)
}

func Pkg(pkg string) string {
	return ExecCollect("go", "list", "-m", pkg)
}

func Version() string {
	v := latestTag()
	if v == "" {
		v = "v0.0.0"
	}
	if !semver.IsValid(v) {
		Exit("VERSION must be in vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]] format")
	}
	parts := breakDown(v)
	switch os.Getenv("VERSION") {
	case "major":
		i, err := strconv.Atoi(strings.TrimPrefix(parts[0], "v"))
		if err != nil {
			Exit("  failed parsing major version", parts[0], err.Error())
		}
		i++
		parts[0] = "v" + strconv.Itoa(i)
	case "minor":
		i, err := strconv.Atoi(parts[1])
		if err != nil {
			Exit("  failed parsing minor version", parts[1], err.Error())
		}
		i++
		parts[1] = strconv.Itoa(i)
	case "patch":
		i, err := strconv.Atoi(parts[2])
		if err != nil {
			Exit("  failed parsing patch version", parts[2], err.Error())
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
	return format(parts)
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
		Exit("VERSION must be in vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]] format")
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

func latestTag() string {
	return ExecCollect("git", "describe", "--abbrev=0")
}

func EnsureRepo(root, repo, dir string) {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			println(">  downloading", dir)
			ExecPlain("git", "clone", repo)
		} else {
			Exit(err.Error())
		}
	} else {
		println(">  updating", dir)
		ExecPlainWithWorkingPath(
			filepath.Join(root, dir),
			"git", "pull",
		)
	}
}
