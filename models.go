package vince

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	APIKey                  []*APIKey
	Subscription            Subscription
	EnterprisePlan          EnterprisePlan
	GoogleAuth              GoogleAuth
	LastSeen                time.Time
	TrialExpiryDate         sql.NullTime
	EmailVerified           bool   `gorm:"not null;default:false"`
	Theme                   string `gorm:"not null;default:system"`
}

type Site struct {
	ID         uint64 `gorm:"primarykey;autoIncrement:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time      `gorm:"index"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
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
	UserID uint64
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
	Scopes                string `gorm:"not null"`
	HourlyAPIRequestLimit uint   `gorm:"not null;default:1000"`
	KeyPrefix             string `gorm:"not null"`
	KeyHash               string `gorm:"not null"`
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

func open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(
		&User{},
		&Site{},
		&EmailVerificationCode{},
		&SpikeNotification{},
		&SiteMembership{},
		&APIKey{},
		&IntroEmail{},
		&CustomDomain{},
		&CreateSiteEmail{},
		&CheckStatEmail{},
		&FeedbackEmail{},
		&Subscription{},
		&SharedLinks{},
		&SentRenewalNotification{},
		&GoogleAuth{},
		&WeeklyReport{},
		&SentWeeklyReport{},
		&MonthlyReport{},
		&SentMonthlyReport{},
		&Goal{},
		&EnterprisePlan{},
		&Invitation{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func closeDB(db *gorm.DB) {
	x, _ := db.DB()
	x.Close()
}

func OpenTest(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := open(filepath.Join(t.TempDir(), "vince.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		r, _ := db.DB()
		r.Close()
	})
	return db
}
