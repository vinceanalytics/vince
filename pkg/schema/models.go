package schema

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
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
	HourlyAPIRequestLimit uint   `gorm:"not null;default:1000"`
	KeyPrefix             string
	KeyHash               string
	UsedAt                time.Time
	Scopes                datatypes.JSONSlice[*Scope]
	ExpiresAt             time.Time
}

func (a *APIKey) ScopeList() (o []string) {
	var b strings.Builder
	for _, r := range a.Scopes {
		for _, v := range r.Verbs {
			b.Reset()
			fmt.Fprintf(&b, "%s:%s", r.Resource, v)
			o = append(o, b.String())
		}
	}
	return
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

type Scope struct {
	Resource Resource
	Verbs    []Verb
}

type Resource uint

const (
	Sites Resource = iota
	Stats
)

var resource_name = map[string]Resource{
	"sites": Sites,
	"stats": Stats,
}

var name_from_resource = map[Resource]string{
	Sites: "sites",
	Stats: "stats",
}

var ErrUnknownResource = errors.New("unknown resource")

func (r *Resource) From(s string) error {
	v, ok := resource_name[s]
	if !ok {
		return ErrUnknownResource
	}
	*r = v
	return nil
}

func (r Resource) String() string {
	return name_from_resource[r]
}

type Verb uint

const (
	All Verb = iota
	Get
	List
	Create
	Update
	Delete
)

var verbs_name = map[string]Verb{
	"*":      All,
	"get":    Get,
	"list":   List,
	"create": Create,
	"update": Update,
	"delete": Delete,
}

var name_from_verb = map[Verb]string{
	All:    "*",
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

type ScopeList datatypes.JSONSlice[*Scope]

func ParseScopes(e ...string) (ScopeList, error) {
	fmt.Println(e)
	m := make(map[Resource]*Scope)
	for _, v := range e {
		p := strings.Split(v, ":")
		var r Resource
		err := r.From(p[0])
		if err != nil {
			return nil, err
		}
		x, ok := m[r]
		if !ok {
			x = &Scope{Resource: r}
			m[r] = x
		}
		switch len(p) {
		case 1:
			x.Verbs = append(x.Verbs, All)
		case 2:
			var a Verb
			err = a.From(p[1])
			if err != nil {
				return nil, err
			}
			x.Verbs = append(x.Verbs, a)
		}
	}
	var ls ScopeList
	for _, v := range m {
		sort.Slice(v.Verbs, func(i, j int) bool {
			return v.Verbs[i] < v.Verbs[j]
		})
		ls = append(ls, v)
	}
	sort.Slice(ls, func(i, j int) bool {
		return ls[i].Resource < ls[j].Resource
	})
	return ls, nil
}

func (ls ScopeList) Can(resource Resource, verb Verb) bool {
	for _, r := range ls {
		if r.Resource == resource {
			for _, v := range r.Verbs {
				if v == All {
					return true
				}
				if v == verb {
					return true
				}
			}
		}
	}
	return false
}
