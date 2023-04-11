package models

import (
	"context"
	"database/sql"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/vince/config"
	"golang.org/x/crypto/bcrypt"
)

type currentUserKey struct{}

type User struct {
	Model
	Name         string
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Sites        []*Site `gorm:"many2many:site_memberships;"`

	EmailVerificationCodes  []*EmailVerificationCode `gorm:"constraint:OnDelete:CASCADE;"`
	IntroEmails             []*IntroEmail            `gorm:"constraint:OnDelete:CASCADE;"`
	FeedbackEmails          []*FeedbackEmail         `gorm:"constraint:OnDelete:CASCADE;"`
	CreateSiteEmails        []*CreateSiteEmail       `gorm:"constraint:OnDelete:CASCADE;"`
	CheckStatEmail          []*CheckStatEmail        `gorm:"constraint:OnDelete:CASCADE;"`
	SentRenewalNotification []*SentRenewalNotification
	APIKeys                 []*APIKey

	Subscription   *Subscription
	EnterprisePlan *EnterprisePlan
	GoogleAuth     *GoogleAuth
	GracePeriod    *GracePeriod

	// for invoice generation and billing
	Organization  string
	PostalAddress string
	VATNumber     string

	LastSeen        time.Time
	TrialExpiryDate sql.NullTime
	EmailVerified   bool `gorm:"not null;default:false"`
	Invitations     []*Invitation
}

func (u *User) Avatar(size int) string {
	q := make(url.Values)
	q.Set("u", strconv.FormatUint(u.ID, 10))
	q.Set("s", strconv.Itoa(size))
	return "/avatar?" + q.Encode()
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

// Load fetches user by uid, preloads Subscription. This returns true if the user
// was found and false otherwise.
func (u *User) Load(ctx context.Context, uid uint64) bool {
	err := Get(ctx).First(u, uid).Error
	if err != nil {
		DBE(ctx, err, "failed to get a user")
		return false
	}
	u.Preload(ctx, "Subscription", "EnterprisePlan")
	return true
}

func (u *User) Preload(ctx context.Context, preload ...string) {
	db := Get(ctx)
	for _, p := range preload {
		db = db.Preload(p)
	}
	err := db.First(u).Error
	if err != nil {
		DBE(ctx, err, "failed to preload "+strings.Join(preload, ","))
	}
}

func (u *User) IsEnterprize(ctx context.Context) bool {
	if u.EnterprisePlan != nil {
		// avoid preloading twice if u was CurrentUser must have been preloaded
		// already.
		return true
	}
	u.Preload(ctx, "EnterprisePlan")
	return u.EnterprisePlan != nil
}

// CountOwnedSites counts sites owned by the user.
func (u *User) CountOwnedSites(ctx context.Context) int64 {
	var o int64
	err := Get(ctx).Model(&Site{}).
		Joins("inner join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' and site_memberships.user_id = ? ", u.ID).
		Count(&o).Error
	if err != nil {
		DBE(ctx, err, "failed to count owned sites")
		return 0
	}
	return o
}

func (u *User) SitesLimit(ctx context.Context) int {
	u.Preload(ctx, "EnterprisePlan")
	x := config.Get(ctx)

	switch {
	case x.IsSelfHost:
		return -1
	case x.IsExempt(u.Email):
		return -1
	case u.EnterprisePlan != nil:
		if u.HasActiveSubscription(ctx) {
			return -1
		}
		return int(x.SiteLimit)
	default:
		return int(x.SiteLimit)
	}
}

func (u *User) HasActiveSubscription(ctx context.Context) bool {
	var count int64
	err := Get(ctx).Model(&Subscription{}).
		Where("user_id = ?", u.ID).
		Where("plan_id", u.EnterprisePlan.PlanID).
		Where("status = ?", "active").Count(&count).Error
	if err != nil {
		DBE(ctx, err, "failed to check active subscription")
		return false
	}
	return count == 1
}

func (u *User) New(r *http.Request) (validation map[string]string, err error) {
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
	if conf.IsSelfHost {
		u.TrialExpiryDate = sql.NullTime{
			Time:  time.Now().AddDate(100, 0, 0),
			Valid: true,
		}
	} else {
		u.TrialExpiryDate = sql.NullTime{
			Time:  time.Now().AddDate(0, 0, 30),
			Valid: true,
		}
	}
	if !conf.EnableEmailVerification {
		u.EmailVerified = true
	}
	return
}

func (u *User) Recipient() string {
	return strings.Split(u.Name, " ")[0]
}

func (u *User) Address() *mail.Address {
	return &mail.Address{
		Name:    u.Name,
		Address: u.Email,
	}
}

func Role(ctx context.Context, uid, sid uint64) (role string) {
	err := Get(ctx).Model(&SiteMembership{}).
		Select("role").
		Where("site_id = ?", sid).
		Where("user_id = ?", uid).
		Limit(1).
		Row().Scan(&role)
	if err != nil {
		DBE(ctx, err, "failed to retrieve role membership")
	}
	return
}
