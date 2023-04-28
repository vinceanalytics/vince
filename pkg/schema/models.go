package schema

import (
	"database/sql"
	"time"
)

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
	IngestRateLimitThreshold    sql.NullInt64

	Users              []*User `gorm:"many2many:site_memberships;"`
	SentWeeklyReports  []*SentWeeklyReport
	SentMonthlyReports []*SentMonthlyReport

	WeeklyReport      *WeeklyReport
	MonthlyReports    *MonthlyReport
	SpikeNotification *SpikeNotification `gorm:"constraint:OnDelete:CASCADE;"`

	Invitations     []*Invitation `gorm:"constraint:OnDelete:CASCADE;"`
	SiteMemberships []*SiteMembership
	SharedLinks     []*SharedLink
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
	UserID uint64 `gorm:"primaryKey"`
	User   *User
	SiteID uint64 `gorm:"primaryKey"`
	Site   *Site
	Role   string `gorm:"not null;default:'owner';check:role in ('owner', 'admin', 'viewer')"`
}

type APIKey struct {
	Model
	UserID                uint64 `gorm:"not null"`
	Name                  string `gorm:"not null"`
	Scopes                string `gorm:"not null;default:stats:read:*"`
	HourlyAPIRequestLimit uint   `gorm:"not null;default:1000"`
	KeyPrefix             string
	UsedAt                time.Time
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

type WeeklyReport struct {
	Model
	SiteID     uint64
	Recipients string
}

type SentWeeklyReport struct {
	Model
	SiteID uint64
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
	Domain    string `gorm:"index"`
	EventName string
	PagePath  string
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
}

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

	Subscription *Subscription

	// for invoice generation and billing
	Organization  string
	PostalAddress string
	VATNumber     string

	LastSeen        time.Time
	TrialExpiryDate sql.NullTime
	EmailVerified   bool `gorm:"not null;default:false"`
	Invitations     []*Invitation
}

type CachedSite struct {
	ID                          uint64
	Domain                      string
	StatsStartDate              time.Time
	IngestRateLimitScaleSeconds uint64
	IngestRateLimitThreshold    sql.NullInt64
	UserID                      uint64
}
