package tests

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/vinceanalytics/vince/tools"
)

var once sync.Once

const driveBinary = "vlg"

// Returns the binary for vince_load_gen
func binary() string {
	once.Do(func() {
		tools.ExecPlainWithWorkingPath(
			tools.RootVince(),
			"go", "install",
			"./tools/vlg",
		)
	})
	return driveBinary
}

func Drive(t *testing.T, script string, f ...func(*exec.Cmd)) string {
	t.Helper()
	var o bytes.Buffer
	cmd := exec.Command(binary(), script)
	cmd.Stderr = &o
	cmd.Stdout = &o
	err := cmd.Run()
	if err != nil {
		t.Fatalf("failed executing driver %s", cmd)
	}
	return strings.TrimSpace(o.String())
}

// Host is like Drive but customize the target host for the driver
func Host(t *testing.T, host, script string) string {
	return Drive(t, script, func(c *exec.Cmd) {
		c.Env = append(c.Env, fmt.Sprintf("VLG_HOST=%s", host))
	})
}
