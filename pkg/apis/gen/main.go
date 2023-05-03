package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gernest/vince/tools"
	_ "k8s.io/code-generator"
)

const (
	rootPackage = "github.com/gernest/vince"
)

var (
	GENERATE_SCRIPT string
	root            string
)

func main() {
	println("### Generating k8s client for crds ###")
	var err error
	root, err = filepath.Abs("../../")
	if err != nil {
		tools.Exit("failed to resolve root", err.Error())
	}
	println(">>> root:", root)
	make()
}

func make() {
	CODEGEN_PKG := tools.ModuleRoot("k8s.io/code-generator")
	println(">>> using codegen: ", CODEGEN_PKG)
	GENERATE_SCRIPT = filepath.Join(CODEGEN_PKG, "generate-groups.sh")
	tools.ExecPlain("chmod", "+x", GENERATE_SCRIPT)
	println("##### Generating site client ######")
	generate("vince", "v1alpha1")
}

func generate(resource string, versions ...string) {
	dir, err := os.MkdirTemp("", resource)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, v := range versions {
		tools.ExecPlain("rm", "-f", filepath.Join(root,
			"/pkg/apis", resource,
			v,
			"zz_generated.deepcopy.go",
		))
	}
	tools.ExecPlain("rm", "-rf", filepath.Join(
		root, "/pkg/gen/client", resource,
	))
	tools.ExecPlain(GENERATE_SCRIPT, "all",
		filepath.Join(rootPackage, "/pkg/gen/client", resource),
		filepath.Join(rootPackage, "/pkg/apis"),
		resource+":"+strings.Join(versions, ","),
		"--go-header-file", filepath.Join(root, "pkg/apis/gen/boilerplate.go.txt"),
		"--output-base", dir,
	)
	tools.ExecPlain("cp", "-r",
		filepath.Join(dir, rootPackage)+"/.",
		root+"/")
}
