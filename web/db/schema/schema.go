package schema

import (
	"time"
)

type Site struct {
	Model
	Domain          string            `gorm:"uniqueIndex" json:"domain"`
	Timezone        string            `gorm:"default:UTC" json:"timezone"`
	Public          bool              `gorm:"not null;default:false" json:"public"`
	Users           []*User           `gorm:"many2many:site_memberships;" json:"-"`
	SiteMemberships []*SiteMembership `json:"siteMemberships,omitempty"`
}

type SiteMembership struct {
	Model
	UserID uint64 `gorm:"primaryKey"`
	User   *User
	SiteID uint64 `gorm:"primaryKey"`
	Site   *Site
	Role   string `gorm:"not null;default:'owner';check:role in ('owner', 'admin', 'viewer')"`
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
	Name         string
	Email        string `gorm:"uniqueIndex"`
	PasswordHash string
	Sites        []*Site `gorm:"many2many:site_memberships;"`
	LastSeen     time.Time
	Invitations  []*Invitation
}
