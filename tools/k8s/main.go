package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/gernest/vince/tools"
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
		tools.ExecCollect("go", "env", "GOPATH"),
		build.Deps[1].Path, build.Deps[1].Version,
	)
	println(">>> using codegen: ", CODEGEN_PKG)
	GENERATE_SCRIPT = filepath.Join(CODEGEN_PKG, "generate-groups.sh")
	tools.ExecPlain("chmod", "+x", GENERATE_SCRIPT)
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
		tools.ExecPlain("rm", "-f", filepath.Join(rootDir,
			"/pkg/apis", resource,
			v,
			"zz_generated.deepcopy.go",
		))
	}
	tools.ExecPlain("rm", "-rf", filepath.Join(
		rootDir, "/pkg/gen/client", resource,
	))
	tools.ExecPlain(GENERATE_SCRIPT, "all",
		filepath.Join(rootPackage, "/pkg/gen/client", resource),
		filepath.Join(rootPackage, "/pkg/apis"),
		resource+":"+strings.Join(versions, ","),
		"--go-header-file", filepath.Join(rootDir, "tools/k8s/boilerplate.go.txt"),
		"--output-base", dir,
	)
	tools.ExecPlain("cp", "-r",
		filepath.Join(dir, rootPackage)+"/.",
		rootDir+"/")
}
