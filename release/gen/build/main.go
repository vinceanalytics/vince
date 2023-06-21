package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vinceanalytics/vince/tools"
)

func main() {
	if os.Getenv("BUILD") != "" {
		sdk := tools.ExecCollect("xcrun", "--show-sdk-path")
		sdk = strings.TrimSpace(sdk)
		sdk = filepath.Join(sdk, "/System/Library/Frameworks")
		println("> using ", sdk)
		tools.ExecPlainWith(
			func(c *exec.Cmd) {
				c.Dir = tools.RootVince()
				c.Env = append(c.Env,
					fmt.Sprintf("FOUNDATION=%s", sdk),
				)
			},
			"goreleaser", "release", "--clean", "--timeout", "60m",
		)
	}
}
