package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	root := tools.RootVince()
	if os.Getenv("DOWNLOAD") != "" {
		download(root)
	}
	if os.Getenv("SITE") != "" {
		tools.ExecPlainWith(func(c *exec.Cmd) {
			c.Dir = root
			c.Env = append(c.Env, "SITE=true")
		},
			"go", "generate", "./website",
		)
	}

}

func download(root string) {
	r := tools.Release(root)
	base := u + r.Meta.Tag
	var b bytes.Buffer
	fmt.Fprintln(&b, "# release", r.Meta.Tag)
	for _, a := range r.Artifacts {
		fmt.Fprintln(&b, "## ", a.ID)
		fmt.Fprintln(&b, `|     filename  |  signature | Size |
| --------------| ----------- | -----|`)
		for _, e := range a.Artifacts {
			fmt.Fprintf(&b, "| %s | %s | `%dMB` |\n", rel(base, e.Name), sig(base, e.Name), e.Extra.Size/(1<<20))
		}
	}
	tools.WriteFile(filepath.Join(root, "website/docs/guide/downloads-table.md"), b.Bytes())
}

const u = `https://github.com/vinceanalytics/vince/releases/download/`

func rel(base, s string) string {
	return fmt.Sprintf("[%s](%s/%s)", s, base, s)
}

func sig(base, s string) string {
	return fmt.Sprintf("[minisig](%s/%s.minisig)", base, s)
}
