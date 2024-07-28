package schema

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Site struct {
	Model
	Domain          string            `gorm:"uniqueIndex" json:"domain"`
	Timezone        string            `gorm:"default:UTC" json:"timezone"`
	Public          bool              `gorm:"not null;default:false" json:"public"`
	Users           []*User           `gorm:"many2many:site_memberships;" json:"-"`
	SiteMemberships []*SiteMembership `json:"siteMemberships,omitempty"`
}

func (s *Site) HasGoals(db *gorm.DB) bool {
	return Exists(db, func(db *gorm.DB) *gorm.DB {
		return db.Model(&Goal{}).Where("domain = ?", s.Domain)
	})
}

func (s *Site) Delete(db *gorm.DB) error {
	return db.Select("SiteMemberships").Delete(s).Error
}

func (s *Site) SafeDomain() string {
	return url.PathEscape(s.Domain)
}

func (s *Site) PreloadSite(db *gorm.DB, preload ...string) error {
	for _, p := range preload {
		db = db.Preload(p)
	}
	err := db.First(s).Error
	if err != nil {
		return fmt.Errorf("preload site%w", err)
	}
	return nil
}

func (s *Site) SiteFor(db *gorm.DB, uid uint64, domain string, roles ...string) error {
	err := db.Model(&Site{}).
		Joins("left join site_memberships on site_memberships.site_id = sites.id").
		Where("site_memberships.user_id = ?", uid).
		Where("sites.domain = ?", domain).
		Where("site_memberships.role IN " + buildRoles(roles...)).First(&s).Error
	if err != nil {
		return fmt.Errorf("find site by owner%w", err)
	}
	return nil
}

func (s *Site) ByDomain(db *gorm.DB, domain string) error {
	return db.Model(&Site{}).
		Where("domain = ?", domain).First(&s).Error
}

func (s *Site) ChangeSiteVisibility(db *gorm.DB, public bool) error {
	return db.Model(s).Update("public", public).Error
}

func buildRoles(roles ...string) string {
	for i := range roles {
		roles[i] = "'" + roles[i] + "'"
	}
	return "(" + strings.Join(roles, ", ") + ")"
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

func (s *SharedLink) Create(db *gorm.DB, sid uint64, name, password string) error {
	id, err := gonanoid.New()
	if err != nil {
		return fmt.Errorf("generating shared link id%w", err)
	}
	s.SiteID = sid
	s.Name = name
	s.Slug = id
	if password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating shared link password%w", err)
		}
		s.PasswordHash = string(b)
	}
	return db.Create(s).Error
}

func (s *SharedLink) Update(db *gorm.DB, name, password string) error {
	if password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating shared link  password%w", err)
		}
		s.PasswordHash = string(b)
	}
	if name != "" {
		s.Name = name
	}
	return db.Save(s).Error
}

func (s *SharedLink) Delete(db *gorm.DB) error {
	return db.Delete(s).Error
}

func (s *SharedLink) URI(base string, site *Site) string {
	query := make(url.Values)
	query.Set("auth", s.Slug)
	return fmt.Sprintf("%s/share/%s?%s", base, url.PathEscape(site.Domain), query.Encode())
}

func (s *SharedLink) ByName(db *gorm.DB, sid uint64, name string) error {
	return db.Model(&SharedLink{}).
		Where("site_id = ?", sid).
		Where("name = ?", name).
		First(s).Error
}

func (s *SharedLink) BySlug(db *gorm.DB, sid uint64, slug string) error {
	return db.Model(&SharedLink{}).
		Where("site_id = ?", sid).
		Where("slug = ?", slug).
		First(s).Error
}

type Goal struct {
	Model
	Domain    string `gorm:"index"`
	EventName string
	PagePath  string
}

var pathRe = regexp.MustCompile(`^\/.*`)
var eventRe = regexp.MustCompile(`^.+`)

func ValidateGoals(event, path string) bool {
	return (event != "" && eventRe.MatchString(event)) ||
		(path != "" && pathRe.MatchString(path))
}

func (g *Goal) Name() string {
	if g.EventName != "" {
		return g.EventName
	}
	return "Visit " + g.PagePath
}

func (g *Goal) Create(db *gorm.DB, domain, event, path string) error {
	// Support multiple goals to be set per site. We have removed unique constraint
	// on goals table, so we perform UPSERT based on the goals fields to avoid
	// creating multiple rows of same goals
	g.Domain = domain
	g.EventName = strings.TrimSpace(event)
	g.PagePath = strings.TrimSpace(path)
	return db.Where(g).FirstOrCreate(&g).Error
}

func (g *Goal) ByEvent(db *gorm.DB, domain, event string) error {
	return db.Model(&Goal{}).Where("domain = ?", domain).
		Where("event_name = ?", event).
		Find(g).Error
}

func (g *Goal) ByPage(db *gorm.DB, domain, page string) error {
	return db.Model(&Goal{}).Where("domain = ?", domain).
		Where("page_path = ?", page).
		Find(&g).Error
}

func Goals(db *gorm.DB, domain string) (o []*Goal, err error) {
	err = db.Model(&Goal{}).Where("domain = ?", domain).Find(&o).Error
	if err != nil {
		err = fmt.Errorf("list goals%w", err)
	}
	return
}

func DeleteGoal(db *gorm.DB, gid, domain string) error {
	id, err := strconv.ParseUint(gid, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing goal id=%q domain=%q%w", gid, domain, err)
	}
	err = db.Where("domain = ?", domain).Delete(&Goal{
		Model: Model{ID: id},
	}).Error
	if err != nil {
		return fmt.Errorf("deleting goal id=%q domain=%q%w", gid, domain, err)
	}
	return nil
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

func (u *User) SetPassword(db *gorm.DB, pwd string) (string, error) {
	if pwd == "" {
		return "is required", nil
	}
	if len(pwd) < 6 {
		return "has to be at least 6 characters", nil
	}
	if len(pwd) > 64 {
		return "cannot be longer than 64 characters", nil
	}
	b, e := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if e != nil {
		return "", fmt.Errorf("hashing password%w", e)

	}
	e = db.Model(u).Update("password_hash", string(b)).Error
	if e != nil {
		return "", fmt.Errorf("update password%w", e)
	}
	return "", nil
}

func (u *User) ByID(db *gorm.DB, uid int64) error {
	return db.First(u, uid).Error
}

func (u *User) NewUser(r *http.Request) (validation map[string]any, err error) {
	u.Name = r.Form.Get("name")
	u.Email = r.Form.Get("email")
	password := r.Form.Get("password")
	passwordConfirm := r.Form.Get("password_confirmation")
	validation = make(map[string]any)
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

func (u *User) Save(db *gorm.DB) error {
	return db.Save(u).Error
}

func (u *User) ByEmail(db *gorm.DB, email string) error {
	return db.Model(&User{}).Where("email = ?", email).First(u).Error
}

func (u *User) PasswordMatch(pwd string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(pwd)) == nil
}

func (u *User) CreateSite(db *gorm.DB, domain string, public bool) error {
	err := db.Model(u).Association("Sites").Append(&Site{
		Domain: domain,
		Public: public,
	})
	if err != nil {
		return fmt.Errorf("create site=%q%w", domain, err)
	}
	return nil
}

func UserIsMember(db *gorm.DB, uid, sid uint64) (bool, error) {
	role, err := Role(db, uid, sid)
	return role != "", err
}

func Role(db *gorm.DB, uid, sid uint64) (role string, err error) {
	err = db.Model(&SiteMembership{}).
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

func (u *User) SiteOwner(db *gorm.DB, sid uint64) error {
	err := db.Model(&User{}).
		Joins("left join site_memberships on site_memberships.user_id = users.id").
		Where("site_memberships.site_id = ?", sid).
		Where("site_memberships.role = ?", "owner").First(u).Error
	if err != nil {
		return fmt.Errorf("finding site owner%w", err)
	}
	return nil
}

func (u *User) CountOwnedSites(db *gorm.DB) (o int64, err error) {
	err = db.Model(&Site{}).
		Joins("inner join  site_memberships on sites.id = site_memberships.site_id and site_memberships.role = 'owner' and site_memberships.user_id = ? ", u.ID).
		Count(&o).Error
	return
}

const emailRegexString = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"

var emailRRe = regexp.MustCompile(emailRegexString)

func validate(field, value, reason string, m map[string]any, f func(string) bool) {
	if f(value) {
		return
	}
	m["validation_"+field] = reason
}

func Exists(g *gorm.DB, where func(db *gorm.DB) *gorm.DB) bool {
	db := where(g).Select("1").Limit(1)
	var n int
	err := db.Find(&n).Error
	return err == nil && n == 1
}
