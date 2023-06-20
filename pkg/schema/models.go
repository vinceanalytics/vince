package schema

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/datatypes"
)

type Site struct {
	Model
	Domain          string `gorm:"uniqueIndex" json:"domain"`
	Timezone        string `gorm:"default:UTC" json:"timezone"`
	Public          bool   `gorm:"not null;default:false" json:"public"`
	Description     string
	StatsStartDate  time.Time `json:"statsStartDate"`
	IngestRateLimit sql.NullFloat64

	UserID uint64

	Invitations []*Invitation `gorm:"constraint:OnDelete:CASCADE;" json:"invitations,omitempty"`
	SharedLinks []*SharedLink `json:"sharedLinks,omitempty"`
}

type EmailVerificationCode struct {
	Model
	Code   uint64
	UserID sql.NullInt64
}

type APIKey struct {
	Model
	UserID                uint64 `gorm:"not null"`
	Name                  string `gorm:"not null"`
	Scopes                string `gorm:"not null;default:stats:read:*"`
	HourlyAPIRequestLimit uint   `gorm:"not null;default:1000"`
	KeyPrefix             string
	KeyHash               string
	UsedAt                time.Time
}

type SharedLink struct {
	Model
	Name         string `gorm:"uniqueIndex;not null"`
	Slug         string `gorm:"uniqueIndex"`
	SiteID       uint64
	PasswordHash string
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
	SiteID uint64 `gorm:"uniqueIndex"`
	UserID uint64 `gorm:"uniqueIndex"`
	Role   string `gorm:"not null;check:role in ('owner', 'admin', 'viewer')"`
}

type Model struct {
	ID        uint64    `gorm:"primarykey;autoIncrement:true" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type User struct {
	Model
	Name                   string `gorm:"uniqueIndex"`
	FullName               string
	Email                  string `gorm:"uniqueIndex"`
	PasswordHash           string
	Sites                  []*Site
	Roles                  []*Role
	EmailVerificationCodes []*EmailVerificationCode `gorm:"constraint:OnDelete:CASCADE;"`
	APIKeys                []*APIKey
	LastSeen               time.Time
	EmailVerified          bool `gorm:"not null;default:false"`
	Invitations            []*Invitation
}

type CachedSite struct {
	ID                          uint64
	Domain                      string
	StatsStartDate              time.Time
	IngestRateLimitScaleSeconds uint64
	IngestRateLimitThreshold    sql.NullInt64
	UserID                      uint64
}

type Verb uint

const (
	Get Verb = iota
	List
	Create
	Update
	Delete
)

var verbs_name = map[string]Verb{
	"get":    Get,
	"list":   List,
	"create": Create,
	"update": Update,
	"delete": Delete,
}

var name_from_verb = map[Verb]string{
	Get:    "get",
	List:   "list",
	Create: "create",
	Update: "update",
	Delete: "delete",
}

var ErrUnknownVerb = errors.New("unknown verb")

func (v *Verb) From(s string) error {
	a, ok := verbs_name[s]
	if !ok {
		return ErrUnknownVerb
	}
	*v = a
	return nil
}

func (v Verb) String() string {
	return name_from_verb[v]
}

type Role struct {
	Model
	UserID  uint64
	Name    string
	Subject string
	Domain  string
	Actions datatypes.JSONSlice[Verb]
}
