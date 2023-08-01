package tests

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"

	"github.com/emersion/go-smtp"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/server"
	"github.com/vinceanalytics/vince/tests/mail"
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
	for _, fn := range f {
		fn(cmd)
	}
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

type Options struct {
	StartSMTP bool
	SMTPOpts  []func(*smtp.Server)
	VinceOpts []func(*config.Options)
}

// Vince configures vince instance and returns context.Context of configured
// values. A lot of vince features support notification via email thats why we
// also return mail.Backend, this email server is not started unless you
// explicitly set o.StartSMTP to true.
//
// Resources are automatically released when the test function is done.
func Vince(t *testing.T, o Options) (context.Context, *mail.Backend) {
	t.Helper()
	vOpts := append(o.VinceOpts, func(o *config.Options) {
		o.DataPath = t.TempDir()
	})
	vo := config.Test(vOpts...)
	be := mail.New(vo.Mailer.SMTP.Address, o.SMTPOpts...)
	ctx, r := server.Configure(context.Background(), vo)
	if o.StartSMTP {
		go be.Start()
		r = append(r, be)
	}
	t.Cleanup(func() {
		r.Close()
	})
	return ctx, be
}
