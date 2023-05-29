package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gernest/vince/tools"
)

func main() {
	sdk := tools.ExecCollect("xcrun", "--show-sdk-path")
	sdk = strings.TrimSpace(sdk)
	sdk = filepath.Join(sdk, "/System/Library/Frameworks")
	println("> using ", sdk)
	tools.ExecPlainWith(
		func(c *exec.Cmd) {
			c.Dir = tools.RootVince()
			c.Env = os.Environ()
			c.Env = append(c.Env,
				fmt.Sprintf("FOUNDATION=%s", sdk),
			)
		},
		"goreleaser", "release", "--clean",
	)
}
