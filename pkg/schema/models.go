package schema

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/core"
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
	UserID          uint64
	Goals           []*Goal
}

type EmailVerificationCode struct {
	Model
	Code   uint64
	UserID sql.NullInt64
}

type APIKey struct {
	Model
	Owner                 string
	Site                  string
	Name                  string `gorm:"not null"`
	HourlyAPIRequestLimit uint   `gorm:"not null;default:1000"`
	KeyPrefix             string
	KeyHash               string `gorm:"index"`
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

func (a *APIKey) Can(ctx context.Context, owner, site string, resource Resource, verb Verb) bool {
	if !a.ExpiresAt.IsZero() && a.ExpiresAt.Before(core.Now(ctx)) {
		// The key has expired
		return false
	}
	if a.Owner != owner {
		return false
	}
	if a.Site != "" && a.Site != site {
		return false
	}
	if len(a.Scopes) == 0 {
		return false
	}
	if !ScopeList(a.Scopes).Can(resource, verb) {
		return false
	}
	return true
}

type Goal struct {
	Model
	SiteID    uint64
	Name      string `gorm:"uniqueIndex"`
	EventName string
	PagePath  string
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
	APIKeys                []*APIKey                `gorm:"foreignKey:Owner;references:Name"`
	LastSeen               time.Time
	EmailVerified          bool `gorm:"not null;default:false"`
}

type Membership struct {
	Model
	UserID uint64
	Site   *Site
	SiteID uint64
	User   *User
	Role   string `gorm:"not null;default:'owner';check:role in ('owner', 'admin', 'viewer')"`
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
