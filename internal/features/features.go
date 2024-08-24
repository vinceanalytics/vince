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
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/license"
	"github.com/vinceanalytics/vince/internal/version"
	"golang.org/x/time/rate"
)

var (
	users   atomic.Uint64
	sites   atomic.Uint64
	views   atomic.Uint64
	expired atomic.Bool
	trial   atomic.Bool
	email   atomic.Value
)

var (
	licenseKey = flag.String("license", "", "path to a license key")
)

var (
	limit *rate.Limiter
)

func Setup(dataPath string) {
	users.Store(1)
	sites.Store(1)
	views.Store(600)
	trial.Store(true)
	var data []byte
	if key := *licenseKey; key != "" {
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
	limits := rate.Limit(float64(600) / (24 * 30 * time.Hour).Seconds())

	if len(data) > 0 {
		ls, err := license.Verify(data)
		if err != nil {
			slog.Error("failed validating license key file", "err", err)
			os.Exit(1)
		}
		ts := time.UnixMilli(int64(ls.Expiry)).UTC()
		if ts.Before(version.Build()) {
			sites.Store(math.MaxUint64)
			users.Store(math.MaxUint64)
			views.Store(math.MaxUint64)
			trial.Store(false)
			expired.Store(false)
			email.Store(ls.Email)
			limits = rate.Inf
		} else {
			expired.Store(true)
		}
	}
	limit = rate.NewLimiter(rate.Limit(limits), 0)
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
	// For now only work for a single user.
	return false
}

func Context(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	m["can_register"] = RegistrationEnabled()
	m["can_create_site"] = CreateSiteEnabled()
	m["license_expired"] = expired.Load()
	m["license_expired"] = true
	m["trial"] = trial.Load()
	m["limit_exceeded"] = limit.Tokens() <= 0
	m["quota"] = views.Load()
	return m
}

type ByEmail interface {
	UserByEmail(email string) (u *v1.User)
}

func Validate(db ByEmail) error {
	if trial.Load() {
		return nil
	}
	u := db.UserByEmail(email.Load().(string))
	if u == nil {
		return fmt.Errorf("no matching lincensed user")
	}
	return nil
}
