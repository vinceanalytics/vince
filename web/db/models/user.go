package models

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gernest/len64/web/db/schema"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type currentUserKey struct{}

type User = schema.User

func SetPassword(ctx context.Context, pwd string) (string, error) {
	if pwd == "" {
		return "is required", nil
	}
	if len(pwd) < 6 {
		return "has to be at least 6 characters", nil
	}
	if len(pwd) > 64 {
		return "cannot be longer than 64 characters", nil
	}
	u := GetUser(ctx)
	b, e := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if e != nil {
		return "", fmt.Errorf("hashing password%w", e)

	}
	e = Get(ctx).Model(u).Update("password_hash", string(b)).Error
	if e != nil {
		return "", fmt.Errorf("update password%w", e)
	}
	return "", nil
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
) error {
	if Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&User{}).Where("email = ?", email)
	}) {
		return nil
	}
	hashPasswd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hashing password%w", err)
	}
	u := &User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashPasswd),
	}
	err = Get(ctx).Create(u).Error
	if err != nil {
		return fmt.Errorf("saving bootstrapped user%w", err)
	}
	return nil
}

func UserByUID(ctx context.Context, uid uint64) (*User, error) {
	var m User
	err := Get(ctx).First(&m, uid).Error
	if err != nil {
		return nil, fmt.Errorf("user by id%w", err)
	}
	return &m, nil
}

// CountOwnedSites counts sites owned by the user.
func CountOwnedSites(ctx context.Context, uid uint64) (int64, error) {
	var o int64
	err := Get(ctx).Model(&Site{}).
		Joins("inner join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' and site_memberships.user_id = ? ", uid).
		Count(&o).Error
	if err != nil {
		return 0, fmt.Errorf("count owned site%w", err)
	}
	return o, nil
}

func NewUser(u *User, r *http.Request) (validation map[string]string, err error) {
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
	return
}

func PasswordMatch(u *User, pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pwd)) == nil
}

func Role(ctx context.Context, uid, sid uint64) (role string, err error) {
	err = Get(ctx).Model(&SiteMembership{}).
		Select("role").
		Where("site_id = ?", sid).
		Where("user_id = ?", uid).
		Limit(1).
		Row().Scan(&role)
	if err != nil {
		err = fmt.Errorf("finding role%w", err)
	}
	return
}

func UserByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := Get(ctx).Model(&User{}).Where("email = ?", email).First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("user by email %w", err)
	}
	return &u, nil
}

func UserByID(ctx context.Context, uid uint64) (*User, error) {
	var m User
	err := Get(ctx).Model(&User{}).Where("id = ?", uid).First(&m).Error
	if err != nil {
		return nil, fmt.Errorf("user by id%w", err)
	}
	return &m, nil
}

func CreateSite(ctx context.Context, usr *User, domain string, public bool) error {
	err := Get(ctx).Model(usr).Association("Sites").Append(&Site{
		Domain: domain,
		Public: public,
	})
	if err != nil {
		return fmt.Errorf("create site=%q%w", domain, err)
	}
	return nil
}
