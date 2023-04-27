package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	_ "k8s.io/code-generator"
)

const (
	rootPackage = "github.com/gernest/vince"
	rootDir     = "."
)

var (
	GENERATE_SCRIPT string
)

func main() {
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	CODEGEN_PKG := fmt.Sprintf("%s/pkg/mod/%s@%s",
		execCollect("go", "env", "GOPATH"),
		build.Deps[0].Path, build.Deps[0].Version,
	)
	println(">>> using codegen: ", CODEGEN_PKG)
	GENERATE_SCRIPT = filepath.Join(CODEGEN_PKG, "generate-groups.sh")
	execPlain("chmod", "+x", GENERATE_SCRIPT)
	println("##### Generating site client ######")
	generate("site", "v1alpha1")
}

func generate(resource string, versions ...string) {
	dir, err := os.MkdirTemp(resource, "")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, v := range versions {
		execPlain("rm", "-f", filepath.Join(rootDir,
			"/pkg/apis", resource,
			v,
			"zz_generated.deepcopy.go",
		))
	}
	execPlain("rm", "-rf", filepath.Join(
		rootDir, "/pkg/gen/client", resource,
	))
	execPlain(GENERATE_SCRIPT, "all",
		filepath.Join(rootPackage, "/pkg/gen/client", resource),
		filepath.Join(rootPackage, "/pkg/apis"),
		resource+":"+strings.Join(versions, ","),
		"--go-header-file", filepath.Join(rootDir, "tools/k8s/boilerplate.go.txt"),
		"--output-base", dir,
	)
	execPlain("cp", "-r",
		filepath.Join(dir, rootPackage)+"/.",
		rootDir+"/")
}

func execCollect(name string, args ...string) string {
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

func execPlain(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
