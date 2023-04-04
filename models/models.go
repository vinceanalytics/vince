package models

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/log"
	"golang.org/x/time/rate"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type currentRoleKey struct{}

func SetRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, currentRoleKey{}, role)
}

func GetRole(ctx context.Context) string {
	if u := ctx.Value(currentRoleKey{}); u != nil {
		return u.(string)
	}
	return ""
}

type siteKey struct{}

func SetSite(ctx context.Context, site *Site) context.Context {
	return context.WithValue(ctx, siteKey{}, site)
}

func GetSite(ctx context.Context) *Site {
	if u := ctx.Value(siteKey{}); u != nil {
		return u.(*Site)
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
	return gp != nil && !gp.Active()
}

func toDate(ts time.Time) time.Time {
	y, m, d := ts.Date()
	return time.Date(
		y, m, d, 0, 0, 0, 0, ts.Location(),
	)
}

type Site struct {
	Model
	Domain                      string `gorm:"uniqueIndex"`
	Description                 string
	Timezone                    string `gorm:"default:UTC"`
	Public                      bool   `gorm:"not null;default:false"`
	StatsStartDate              time.Time
	HasStats                    bool   `gorm:"not null,default:false"`
	Locked                      bool   `gorm:"not null,default:false"`
	IngestRateLimitScaleSeconds uint64 `gorm:"not null;default:60"`
	IngestRateLimitThreshold    uint64

	Users              []*User `gorm:"many2many:site_memberships;"`
	SentWeeklyReports  []*SentWeeklyReport
	SentMonthlyReports []*SentMonthlyReport

	WeeklyReport      WeeklyReport
	MonthlyReports    MonthlyReport
	GoogleAuth        GoogleAuth        `gorm:"constraint:OnDelete:CASCADE;"`
	CustomDomain      CustomDomain      `gorm:"constraint:OnDelete:CASCADE;"`
	SpikeNotification SpikeNotification `gorm:"constraint:OnDelete:CASCADE;"`

	Invitations []*Invitation `gorm:"constraint:OnDelete:CASCADE;"`
	SharedLinks []*SharedLink
}

func (s *Site) IsMember(userId uint64) bool {
	for _, m := range s.Users {
		if m.ID == userId {
			return true
		}
	}
	return false
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
	UserID    uint64 `gorm:"primaryKey"`
	SiteID    uint64 `gorm:"primaryKey"`
	Role      string `gorm:"not null;default:'owner';check:role in ('owner', 'admin', 'viewer')"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
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

func (ak *APIKey) RateLimit() (uint64, rate.Limit, int) {
	r := rate.Limit(float64(ak.HourlyAPIRequestLimit) / time.Hour.Seconds())
	return ak.ID, r, 10
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
	PlanID         uint64    `gorm:"not null"`
	UpdateURL      string    `gorm:"not null"`
	CancelURL      string    `gorm:"not null"`
	Status         string    `gorm:"not null;check:status in ('active', 'past_due', 'deleted', 'paused')"`
	NextBillAmount string    `gorm:"not null"`
	NextBillDate   time.Time `gorm:"not null"`
	LastBillDate   time.Time
}

func (sub *Subscription) GetEnterPrise(ctx context.Context) *EnterprisePlan {
	var e EnterprisePlan
	err := Get(ctx).Model(&EnterprisePlan{}).
		Where("plan_id = ? ", sub.PlanID).
		Where("user_id = ?", sub.UserID).First(&e).Error
	if err != nil {
		DBE(ctx, err, "failed getting enterprise plan from subscription")
		return nil
	}
	return &e
}

func DBE(ctx context.Context, err error, msg string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}
	log.Get(ctx).Err(err).Msg(msg)
}

type SharedLink struct {
	Model
	Name         string `gorm:"uniqueIndex;not null"`
	Slug         string `gorm:"uniqueIndex"`
	SiteID       uint64
	PasswordHash string
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
	PlanID                uint64 `gorm:"not null"`
	UserID                uint64 `gorm:"not null;uniqueIndex"`
	BillingInterval       string `gorm:"not null;check:billing_interval in ('monthly', 'yearly')"`
	MonthlyPageViewLimit  uint64 `gorm:"not null"`
	HourlyAPIRequestLimit uint64 `gorm:"not null"`
	SiteLimit             uint64 `gorm:"default:50"`
}

type Invitation struct {
	Model
	Email  string `gorm:"not null;uniqueIndex"`
	SiteID int    `gorm:"uniqueIndex"`
	UserID uint64 `gorm:"uniqueIndex"`
	Role   string `gorm:"not null;check:role in ('owner', 'admin', 'viewer')"`
}

type Model struct {
	ID        uint64 `gorm:"primarykey;autoIncrement:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func Database(cfg *config.Config) string {
	return filepath.Join(cfg.DataPath, "vince.db")
}

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.SetupJoinTable(&User{}, "Sites", &SiteMembership{})
	db.SetupJoinTable(&Site{}, "Users", &SiteMembership{})
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
		&SharedLink{},
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
	generateCode := !exists(db, func(db *gorm.DB) *gorm.DB {
		return db.Model(&EmailVerificationCode{}).Where("code <> 0")
	})
	if generateCode {
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

func Exists(ctx context.Context, where func(db *gorm.DB) *gorm.DB) bool {
	return exists(Get(ctx), where)
}

func exists(db *gorm.DB, where func(db *gorm.DB) *gorm.DB) bool {
	db = where(db).Select("1").Limit(1)
	var n int
	err := db.Find(&n).Error
	return err == nil && n == 1
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
