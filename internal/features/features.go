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
	"errors"
	"sync/atomic"
	"time"

	v1 "github.com/vinceanalytics/vince/gen/go/vince/v1"
	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/version"
)

var (
	Expires atomic.Uint64
	valid   atomic.Bool
	Email   atomic.Value
)

var (
	ErrExpired = errors.New("license expired")
	ErrEmail   = errors.New("invalid admin email")
)

func Apply(ls *v1.License) {
	Expires.Store(ls.Expiry)
	Email.Store(ls.Email)
}

func Validate() error {
	return IsValid(Email.Load().(string), Expires.Load())
}

func Valid() (ok bool) {
	ok = IsValid(Email.Load().(string), Expires.Load()) == nil
	valid.Store(ok)
	return
}

func IsValid(email string, expiry uint64) error {
	best := time.UnixMilli(int64(expiry)).UTC()
	build := version.Build()
	if best.Before(build) {
		// valid license but wrong build
		return ErrExpired
	}
	if config.C.Admin.Email != email {
		return ErrEmail
	}
	return nil
}

func Context(m map[string]any) map[string]any {
	if m == nil {
		m = make(map[string]any)
	}
	m["valid_license"] = valid.Load()
	return m
}
