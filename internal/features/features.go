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
	"flag"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"github.com/vinceanalytics/vince/internal/license"
	"golang.org/x/time/rate"
)

var (
	users   atomic.Uint64
	sites   atomic.Uint64
	views   atomic.Uint64
	expires atomic.Uint64
	expired atomic.Bool
	trial   atomic.Bool
)

var (
	licenseKey         = flag.String("license", "", "path to a license key")
	enableRegistration = flag.Bool("enable-registration", false, "allows registering new users")
)

var (
	limit rate.Limiter
)

func init() {
	users.Store(1)
	sites.Store(1)
	views.Store(600)
	trial.Store(true)
	expires.Store(uint64(time.Now().UTC().Add(24 * 30 * time.Hour).UnixMilli()))
	if key := *licenseKey; key != "" {
		data, err := os.ReadFile(key)
		if err != nil {
			slog.Error("failed reading license key file", "path", key, "err", err)
			os.Exit(1)
		}
		ls, err := license.Verify(data)
		if err != nil {
			slog.Error("failed validating license key file", "path", key, "err", err)
			os.Exit(1)
		}
		sites.Store(ls.Sites)
		users.Store(ls.Users)
		views.Store(ls.Views)
		expires.Store(ls.Expiry)
		trial.Store(false)
	}

	limits := float64(views.Load()) / (24 * 30 * time.Hour).Seconds()
	limit = *rate.NewLimiter(rate.Limit(limits), 0)
}

// Allow accepts an event request. This applies to events sent for valid
// registered sites and is measured across all sites.
func Allow() bool {
	return !expired.Load() &&
		limit.Allow()
}

func CreateSiteEnabled() bool {
	return !expired.Load() &&
		sites.Load() > 0
}

func RegistrationEnabled() bool {
	return *enableRegistration &&
		!expired.Load() &&
		users.Load() > 0
}

func Context(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	m["can_register"] = RegistrationEnabled()
	m["can_create_site"] = CreateSiteEnabled()
	m["license_expired"] = expired.Load()
	m["trial"] = trial.Load()
	m["exceed_quota"] = limit.Tokens() <= 0
	return m
}
