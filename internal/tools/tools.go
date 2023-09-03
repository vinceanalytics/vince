package tools

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/semver"
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

func WriteFile(path string, data []byte) {
	err := os.WriteFile(path, data, 0600)
	if err != nil {
		Exit("failed to write ", path, err.Error())
	}
	println("    write: ", path)
}

func MkDir(path ...string) {
	ExecPlain("mkdir", "-pv", filepath.Join(path...))
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
	ExecPlain("rm", "-rfv", path)
}

func CopyFile(src, dest string) {
	ExecPlain("cp", "-v", src, dest)
}

func Copy(dest string, src ...string) {
	ExecPlain("cp", "-fv", filepath.Join(src...), dest)
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
