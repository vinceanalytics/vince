// Package features exposes features allowed on the vince instance. While I am
// grateful to the open source community , I also need to sustain myself and ensure
// continuity of vince development.
//
// We only allow up to 30 day trial without a license, the trial is also limited
// in scope to avoid automated reload after every 30 days.
//
// All code is available and people are welcome to fork and remove the
// restrictions, just remember this, I solely rely on remote work and lately I
// can't secure employment because REMOTE nowadays means REMOTE US or REMOTE
// EUROPE.
//
// I am trying to make a honest living and I don't like handouts. A lot of
// research and novel engineering work has gone into this project to ensure I
// provide unique value to my users.
//
// You are welcome.
package features

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/license"
	"github.com/vinceanalytics/vince/internal/version"
)

var (
	expired atomic.Bool
	trial   atomic.Bool
	email   atomic.Value
)

func Setup(dataPath string) {
	trial.Store(true)
	var data []byte
	if key := config.C.License; key != "" {
		// try reading  license key
		var err error
		data, err = os.ReadFile(key)
		if err != nil {
			slog.Error("failed reading license key file", "path", key, "err", err)
			os.Exit(1)
		}
	} else {
		// A usermight have applied a license on the UI try reading it from our data path
		data, _ = os.ReadFile(filepath.Join(dataPath, "LICENSE"))
	}

	if len(data) > 0 {
		ls, err := license.Verify(data)
		if err != nil {
			slog.Error("failed validating license key file", "err", err)
			os.Exit(1)
		}
		ts := time.UnixMilli(int64(ls.Expiry)).UTC()
		if ts.Before(version.Build()) {
			trial.Store(false)
			expired.Store(false)
			email.Store(ls.Email)
		} else {
			expired.Store(true)
		}
	}
}

func Context(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	m["license_expired"] = expired.Load()
	m["trial"] = trial.Load()
	return m
}

func Validate() error {
	if trial.Load() {
		return nil
	}
	if config.C.Admin.Email != email.Load().(string) {
		return fmt.Errorf("no matching lincensed user")
	}
	return nil
}
