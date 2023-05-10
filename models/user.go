package models

import (
	"context"
	"net/http"
	"strings"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/schema"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type currentUserKey struct{}

type User = schema.User

func SetPassword(ctx context.Context, pwd string) (err string) {
	if pwd == "" {
		return "is required"
	}
	if len(pwd) < 6 {
		return "has to be at least 6 characters"
	}
	if len(pwd) > 64 {
		return "cannot be longer than 64 characters"
	}
	u := GetUser(ctx)
	b, e := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if e != nil {
		// It is fatal if we can no longer hash passwords terminate the process
		log.Get(ctx).Fatal().AnErr("err", e).Msg("failed hashing password")
	}
	e = Get(ctx).Model(u).Update("password_hash", string(b)).Error
	if e != nil {
		LOG(ctx, e, "failed to update user password")
		return "server error"
	}
	return ""
}
func SetUser(ctx context.Context, usr *User) context.Context {
	return context.WithValue(ctx, currentUserKey{}, usr)
}

func GetUser(ctx context.Context) *User {
	if u := ctx.Value(currentUserKey{}); u != nil {
		return u.(*User)
	}
	return nil
}

// Create a user and API key. No validation is done. If A User exists no further
// operations are done.
func Bootstrap(
	ctx context.Context,
	name, email, password, key string,
) {
	if Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&User{}).Where("email = ?", email)
	}) {
		return
	}
	hashPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Get(ctx).Fatal().Err(err).Msg("failed to hash password")
	}
	hash, prefix := ProcessAPIKey(ctx, key)
	u := User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashPasswd),
		APIKeys: []*APIKey{
			{
				Name:      name,
				KeyPrefix: prefix,
				KeyHash:   hash,
			},
		},
	}
	if config.Get(ctx).EnableEmailVerification {
		u.EmailVerified = true
	}
	err = Get(ctx).Create(u).Error
	if err != nil {
		log.Get(ctx).Fatal().Err(err).Msg("failed to save bootstrapped user")
	}
}

func UserByUID(ctx context.Context, uid uint64) (u *User) {
	var m User
	err := Get(ctx).First(&m, uid).Error
	if err != nil {
		LOG(ctx, err, "failed to get a user")
		return
	}
	PreloadUser(ctx, &m, "Subscription")
	return &m
}

func PreloadUser(ctx context.Context, u *User, preload ...string) {
	db := Get(ctx)
	for _, p := range preload {
		db = db.Preload(p)
	}
	err := db.First(u).Error
	if err != nil {
		LOG(ctx, err, "failed to preload "+strings.Join(preload, ","))
	}
}

// CountOwnedSites counts sites owned by the user.
func CountOwnedSites(ctx context.Context, uid uint64) int64 {
	var o int64
	err := Get(ctx).Model(&Site{}).
		Joins("inner join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' and site_memberships.user_id = ? ", uid).
		Count(&o).Error
	if err != nil {
		LOG(ctx, err, "failed to count owned sites")
		return 0
	}
	return o
}

func NewUser(u *User, r *http.Request) (validation map[string]string, err error) {
	conf := config.Get(r.Context())
	u.Name = r.Form.Get("name")
	u.Email = r.Form.Get("email")
	password := r.Form.Get("password")
	passwordConfirm := r.Form.Get("password_confirmation")
	validation = make(map[string]string)
	validate("name", u.Name, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("email", u.Email, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("email", u.Email, "invalid email", validation, func(s string) bool {
		return emailRRe.MatchString(s)
	})
	validate("password", password, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("password", password, "has to be at least 6 characters", validation, func(s string) bool {
		return len(s) >= 6
	})
	validate("password", password, "cannot be longer than 64 characters", validation, func(s string) bool {
		return len(s) <= 64
	})
	validate("password_confirmation", passwordConfirm, "required", validation, func(s string) bool {
		return s != ""
	})
	validate("password_confirmation", passwordConfirm, "must match password", validation, func(s string) bool {
		return s == password
	})
	if len(validation) != 0 {
		return
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.PasswordHash = string(b)
	if !conf.EnableEmailVerification {
		u.EmailVerified = true
	}
	return
}

func PasswordMatch(u *User, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pwd)) == nil
}

func Role(ctx context.Context, uid, sid uint64) (role string) {
	err := Get(ctx).Model(&SiteMembership{}).
		Select("role").
		Where("site_id = ?", sid).
		Where("user_id = ?", uid).
		Limit(1).
		Row().Scan(&role)
	if err != nil {
		LOG(ctx, err, "failed to retrieve role membership")
	}
	return
}

func UserByEmail(ctx context.Context, email string) *User {
	var u User
	err := Get(ctx).Model(&User{}).Where("email = ?", email).First(&u).Error
	if err != nil {
		LOG(ctx, err, "failed to get user by email")
		return nil
	}
	return &u
}

func UserByID(ctx context.Context, uid uint64) (u *User) {
	var m User
	err := Get(ctx).Model(&User{}).Where("id = ?", uid).First(&m).Error
	if err != nil {
		LOG(ctx, err, "failed to get user by id", func(e *zerolog.Event) *zerolog.Event {
			return e.Uint64("uid", uid)
		})
		return
	}
	return &m
}

func CreateSite(ctx context.Context, usr *User, domain string, public bool) bool {
	err := Get(ctx).Model(usr).Association("Sites").Append(&Site{
		Domain: domain,
		Public: public,
	})
	if err != nil {
		LOG(ctx, err, "failed to create a new site", func(e *zerolog.Event) *zerolog.Event {
			return e.Str("domain", domain).Bool("public", public)
		})
		return false
	}
	return true
}
