package models

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"net/http"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type currentUserKey struct{}

type User struct {
	Model
	Name         string
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Sites        []*Site

	SiteMembership          []*SiteMembership
	EmailVerificationCodes  []*EmailVerificationCode `gorm:"constraint:OnDelete:CASCADE;"`
	IntroEmail              []*IntroEmail            `gorm:"constraint:OnDelete:CASCADE;"`
	FeedbackEmail           []*FeedbackEmail         `gorm:"constraint:OnDelete:CASCADE;"`
	CreateSiteEmails        []*CreateSiteEmail       `gorm:"constraint:OnDelete:CASCADE;"`
	CheckStatEmail          []*CheckStatEmail        `gorm:"constraint:OnDelete:CASCADE;"`
	SentRenewalNotification []*SentRenewalNotification
	APIKeys                 []*APIKey
	Subscription            Subscription
	EnterprisePlan          EnterprisePlan
	GoogleAuth              GoogleAuth
	LastSeen                time.Time
	TrialExpiryDate         sql.NullTime
	EmailVerified           bool   `gorm:"not null;default:false"`
	Theme                   string `gorm:"not null;default:system"`
	GracePeriod             *GracePeriod
	Invitations             []*Invitation
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

func SetCurrentUser(ctx context.Context, usr *User) context.Context {
	return context.WithValue(ctx, currentUserKey{}, usr)
}

func GetCurrentUser(ctx context.Context) *User {
	if u := ctx.Value(currentUserKey{}); u != nil {
		return u.(*User)
	}
	return nil
}

type GracePeriod struct {
	Model
	UserID             uint64
	EndDate            time.Time
	AllowanceRRequired uint
	IsOver             bool
	ManualLock         bool
}

func (gp *GracePeriod) Start(allowance uint) {
	gp.EndDate = toDate(time.Now()).AddDate(0, 0, 7)
	gp.AllowanceRRequired = allowance
	gp.IsOver = false
	gp.ManualLock = false
}

func (gp *GracePeriod) StartManualLock(allowance uint) {
	gp.EndDate = time.Time{}
	gp.AllowanceRRequired = allowance
	gp.IsOver = false
	gp.ManualLock = true
}

func (gp *GracePeriod) End() {
	gp.IsOver = true
}

func (gp *GracePeriod) Active() bool {
	return gp != nil && (gp.ManualLock || gp.EndDate.After(toDate(time.Now())))
}

func (gp *GracePeriod) Expired() bool {
	return gp == nil || !gp.Active()
}

func toDate(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(
		y, m, d, 0, 0, 0, 0, ts.Location(),
	)
}

type Site struct {
	Model
	UserID     uint64
	Domain     string     `gorm:"uniqueIndex"`
	Timezone   string     `gorm:"default:UTC"`
	Public     bool       `gorm:"not null;default:false"`
	GoogleAuth GoogleAuth `gorm:"constraint:OnDelete:CASCADE;"`

	SiteMembership     []*SiteMembership `gorm:"constraint:OnDelete:CASCADE;"`
	WeeklyReport       WeeklyReport
	SentWeeklyReports  []*SentWeeklyReport
	MonthlyReports     *MonthlyReport
	SentMonthlyReports []*SentMonthlyReport

	SpikeNotifications          []*SpikeNotification `gorm:"constraint:OnDelete:CASCADE;"`
	IngestRateLimitScaleSeconds uint64               `gorm:"not null;default:60"`
	IngestRateLimitThreshold    uint64

	StatsStartDate time.Time
	HasStats       bool `gorm:"not null,default:false"`
	Locked         bool `gorm:"not null,default:false"`

	CustomDomains []*CustomDomain `gorm:"constraint:OnDelete:CASCADE;"`
	Invitations   []*Invitation   `gorm:"constraint:OnDelete:CASCADE;"`
}

type CustomDomain struct {
	Model
	Domain             string `gorm:"not null"`
	SiteID             uint64 `gorm:"uniqueIndex"`
	HasSSLCertificates bool   `gorm:"not null,default:false"`
}

type CreateSiteEmail struct {
	Model
	UserID uint64
}

type CheckStatEmail struct {
	Model
	UserID uint64
}

type EmailVerificationCode struct {
	Model
	Code   uint64
	UserID sql.NullInt64
}

type SpikeNotification struct {
	Model
	SiteID     uint64 `gorm:"uniqueIndex"`
	Threshold  uint
	LastSent   time.Time
	Recipients string
}

type SiteMembership struct {
	Model
	UserID uint64
	SiteID uint64
	Role   string `gorm:"not null;default:owner"`
}

type APIKey struct {
	Model
	UserID                uint64 `gorm:"not null"`
	Name                  string `gorm:"not null"`
	Scopes                string `gorm:"not null;default:stats:read:*"`
	HourlyAPIRequestLimit uint   `gorm:"not null;default:1000"`
	KeyPrefix             string `gorm:"not null"`
	KeyHash               string `gorm:"not null"`
}

func (ak *APIKey) New(ctx context.Context) {
	b := make([]byte, 64)
	_, err := crand.Read(b)
	if err != nil {
		// something is really wrong when we cant generate random data.
		log.Get(ctx).Fatal().Msg("failed to generate random data " + err.Error())
	}
	key := base64.StdEncoding.EncodeToString(b)[0:64]
	ak.KeyHash = HashAPIKey(ctx, key)
	ak.KeyPrefix = key[0:6]
}

func HashAPIKey(ctx context.Context, key string) string {
	h := sha256.New()
	h.Write([]byte(config.Get(ctx).SecretKeyBase))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}

type IntroEmail struct {
	Model
	UserID uint64
}

type FeedbackEmail struct {
	Model
	UserID uint64
}

type Subscription struct {
	Model
	UserID         int
	SubID          string    `gorm:"uniqueIndex"`
	PlanID         string    `gorm:"not null"`
	UpdateURL      string    `gorm:"not null"`
	CancelURL      string    `gorm:"not null"`
	Status         string    `gorm:"not null"`
	CurrencyCode   string    `gorm:"not null:default:USD"`
	NextBillAmount string    `gorm:"not null"`
	NextBillDate   time.Time `gorm:"not null"`
	LastBillDate   time.Time
}

type SharedLinks struct {
	Model
	Name string `gorm:"not null"`
}

type SentRenewalNotification struct {
	Model
	UserID uint64
}

type GoogleAuth struct {
	Model
	UserID       uint64
	SiteID       uint64
	Email        string    `gorm:"not null"`
	RefreshToken string    `gorm:"not null"`
	AccessToken  string    `gorm:"not null"`
	Property     string    `gorm:"not null"`
	Expires      time.Time `gorm:"not null"`
}

type WeeklyReport struct {
	Model
	SiteID int
	Email  string
}

type SentWeeklyReport struct {
	Model
	SiteID int
	Year   int
	Week   int
}

type MonthlyReport struct {
	Model
	SiteID uint64
	Email  string
}

type SentMonthlyReport struct {
	Model
	SiteID uint64
	Year   int
	Week   int
}

type Goal struct {
	Model
	Domain    string `gorm:"uniqueIndex"`
	EventName string
	PagePath  string
}

type EnterprisePlan struct {
	Model
	UserID                uint64 `gorm:"not null;uniqueIndex"`
	BillingInterval       string `gorm:"not null"`
	MonthlyPageViewLimit  uint64 `gorm:"not null"`
	HourlyAPIRequestLimit uint64 `gorm:"not null"`
	SiteLimit             uint64 `gorm:"default:50"`
}

type Invitation struct {
	Model
	Email  string `gorm:"not null;uniqueIndex"`
	SiteID int    `gorm:"uniqueIndex"`
	UserID uint64 `gorm:"uniqueIndex"`
	Role   string `gorm:"not null"`
}

type Model struct {
	ID        uint64 `gorm:"primarykey;autoIncrement:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&APIKey{},
		&CheckStatEmail{},
		&CreateSiteEmail{},
		&CustomDomain{},
		&EmailVerificationCode{},
		&EnterprisePlan{},
		&FeedbackEmail{},
		&Goal{},
		&GoogleAuth{},
		&GracePeriod{},
		&IntroEmail{},
		&Invitation{},
		&MonthlyReport{},
		&SentMonthlyReport{},
		&SentRenewalNotification{},
		&SentWeeklyReport{},
		&SharedLinks{},
		&SiteMembership{},
		&Site{},
		&SpikeNotification{},
		&Subscription{},
		&User{},
		&WeeklyReport{},
	)
	if err != nil {
		return nil, err
	}
	// fill verification codes
	var codes int64
	if err = db.Model(&EmailVerificationCode{}).Count(&codes).Error; err != nil {
		return nil, err
	}
	if codes == 0 {
		// We generate end-start rows on email_verification_codes table with random
		// code between start and end. user_id is set to null.
		start := 1000
		end := 9999
		batch := make([]*EmailVerificationCode, 0, end-start)
		for i := start; i < end; i += 1 {
			batch = append(batch, &EmailVerificationCode{
				Code: uint64(i),
			})
		}
		rand.Shuffle(len(batch), func(i, j int) {
			batch[i], batch[j] = batch[j], batch[i]
		})
		err = db.CreateInBatches(batch, 100).Error
		if err != nil {
			return nil, err
		}

	}
	return db, nil
}

func CloseDB(db *gorm.DB) {
	x, _ := db.DB()
	x.Close()
}

func OpenTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "vince.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		r, _ := db.DB()
		r.Close()
	})
	return db
}

type dbKey struct{}

func Set(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, db)
}

func Get(ctx context.Context) *gorm.DB {
	return ctx.Value(dbKey{}).(*gorm.DB)
}

// Check performs health check on the database. This make sure we can query the
// database
func Check(ctx context.Context) bool {
	return Get(ctx).Exec("SELECT 1").Error == nil
}

const emailRegexString = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"

var emailRRe = regexp.MustCompile(emailRegexString)

func validate(field, value, reason string, m map[string]string, f func(string) bool) {
	if f(value) {
		return
	}
	m[field] = reason
}
