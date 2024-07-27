package models

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gernest/len64/web/db/schema"
	"github.com/glebarez/sqlite"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

type Site = schema.Site

func UpdateSiteStartDate(ctx context.Context, sid uint64, start time.Time) error {
	err := Get(ctx).Model(&Site{}).Where("id = ?", sid).Update("stats_start_date", start).Error
	if err != nil {
		return fmt.Errorf("update stats_start_date%w", err)
	}
	return nil
}

func DeleteSite(ctx context.Context, site *Site) error {
	err := Get(ctx).Select("SiteMemberships").Delete(site).Error
	if err != nil {
		return fmt.Errorf("delete site%w", err)
	}
	return nil
}

func SafeDomain(s *Site) string {
	return url.PathEscape(s.Domain)
}

func PreloadSite(ctx context.Context, u *Site, preload ...string) error {
	db := Get(ctx)
	for _, p := range preload {
		db = db.Preload(p)
	}
	err := db.First(u).Error
	if err != nil {
		return fmt.Errorf("preload site%w", err)
	}
	return nil
}

var domainRe = regexp.MustCompile(`(?P<domain>(?:[a-z0-9]+(?:-[a-z0-9]+)*\.)+[a-z]{2,})`)

func ValidateSiteDomain(ctx context.Context, domain string) (good, bad string) {
	good = CleanupDOmain(domain)
	if good == "" {
		bad = "is required"
		return
	}
	if !domainRe.MatchString(good) {
		bad = "only letters, numbers, slashes and period allowed"
		return
	}
	if strings.ContainsAny(domain, reservedChars) {
		bad = "must not contain URI reserved characters " + reservedChars
		return
	}
	if Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&Site{}).Where("domain = ?", domain)
	}) {
		bad = " already exists"
	}
	return
}

const reservedChars = `:?#[]@!$&'()*+,;=`

func CleanupDOmain(domain string) string {
	domain = strings.TrimSpace(domain)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "www.")
	domain = strings.TrimSuffix(domain, "/")
	domain = strings.ToLower(domain)
	return domain
}

func SiteOwner(ctx context.Context, sid uint64) (*User, error) {
	var u User
	err := Get(ctx).Model(&User{}).
		Joins("left join site_memberships on site_memberships.user_id = users.id").
		Where("site_memberships.site_id = ?", sid).
		Where("site_memberships.role = ?", "owner").First(&u).Error
	if err != nil {
		return nil, fmt.Errorf("finding site owner%w", err)
	}
	return &u, nil
}

func SiteFor(ctx context.Context, uid uint64, domain string, roles ...string) (*Site, error) {
	var site Site
	err := Get(ctx).Model(&Site{}).
		Joins("left join site_memberships on site_memberships.site_id = sites.id").
		Where("site_memberships.user_id = ?", uid).
		Where("sites.domain = ?", domain).
		Where("site_memberships.role IN " + buildRoles(roles...)).First(&site).Error
	if err != nil {
		return nil, fmt.Errorf("find site by owner%w", err)
	}
	return &site, nil
}

func buildRoles(roles ...string) string {
	for i := range roles {
		roles[i] = "'" + roles[i] + "'"
	}
	return "(" + strings.Join(roles, ", ") + ")"
}

func SiteHasGoals(ctx context.Context, domain string) bool {
	return Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&Goal{}).Where("domain = ?", domain)
	})
}

func SiteByDomain(ctx context.Context, domain string) (*Site, error) {
	var s Site
	err := Get(ctx).Model(&Site{}).Where("domain = ?", domain).First(&s).Error
	if err != nil {
		return nil, fmt.Errorf("site by domain%w", err)
	}
	return &s, nil
}

func ChangeSiteVisibility(ctx context.Context, site *Site, public bool) error {
	err := Get(ctx).Model(site).Update("public", public).Error
	if err != nil {
		return fmt.Errorf("change site visibility%w", err)
	}
	return nil
}

func UserIsMember(ctx context.Context, uid, sid uint64) (bool, error) {
	role, err := Role(ctx, uid, sid)
	return role != "", err
}

type SiteMembership = schema.SiteMembership

type SharedLink = schema.SharedLink

func CreateSharedLink(ctx context.Context, sid uint64, name, password string) (*SharedLink, error) {
	id, err := gonanoid.New()
	if err != nil {
		return nil, fmt.Errorf("generating shared link id%w", &err)
	}
	shared := &SharedLink{
		SiteID: sid,
		Name:   name,
		Slug:   id,
	}
	if password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("generating shared link password%w", &err)
		}
		shared.PasswordHash = string(b)
	}
	err = Get(ctx).Create(shared).Error
	if err != nil {
		return nil, fmt.Errorf("creating shared link %w", &err)
	}
	return shared, nil
}

func UpdateSharedLink(ctx context.Context, shared *SharedLink, name, password string) error {
	if password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("generating shared link  password%w", &err)
		}
		shared.PasswordHash = string(b)
	}
	if name != "" {
		shared.Name = name
	}
	err := Get(ctx).Save(shared).Error
	if err != nil {
		return fmt.Errorf("updating shared link %w", &err)
	}
	return nil
}

func SharedLinkURL(base string, site *Site, link *SharedLink) string {
	query := make(url.Values)
	query.Set("auth", link.Slug)
	return fmt.Sprintf("%s/share/%s?%s", base, url.PathEscape(site.Domain), query.Encode())
}

func GetSharedLink(ctx context.Context, sid uint64, name string) (*SharedLink, error) {
	var shared SharedLink
	err := Get(ctx).Model(&SharedLink{}).
		Where("site_id = ?", sid).
		Where("name = ?", name).
		First(&shared).Error
	if err != nil {
		return nil, fmt.Errorf("finding shared  link site_id%d name=%q%w", sid, name, &err)
	}
	return &shared, nil
}

func GetSharedLinkWithSlug(ctx context.Context, sid uint64, slug string) (*SharedLink, error) {
	var shared SharedLink
	err := Get(ctx).Model(&SharedLink{}).
		Where("site_id = ?", sid).
		Where("slug = ?", slug).
		First(&shared).Error
	if err != nil {
		return nil, fmt.Errorf("finding shared  link site_id%d slug=%q%w", sid, slug, err)
	}
	return &shared, nil
}

func DeleteSharedLink(ctx context.Context, shared *SharedLink) error {
	err := Get(ctx).Delete(shared).Error
	if err != nil {
		return fmt.Errorf("deleting  shared  link %w", err)
	}
	return nil
}

type Goal = schema.Goal

var pathRe = regexp.MustCompile(`^\/.*`)
var eventRe = regexp.MustCompile(`^.+`)

func ValidateGoals(event, path string) bool {
	return (event != "" && eventRe.MatchString(event)) ||
		(path != "" && pathRe.MatchString(path))
}

func GoalName(g *Goal) string {
	if g.EventName != "" {
		return g.EventName
	}
	return "Visit " + g.PagePath
}

func CreateGoal(ctx context.Context, domain, event, path string) (*Goal, error) {
	// Support multiple goals to be set per site. We have removed unique constraint
	// on goals table, so we perform UPSERT based on the goals fields to avoid
	// creating multiple rows of same goals
	var o Goal
	err := Get(ctx).Where(&Goal{
		Domain:    domain,
		EventName: strings.TrimSpace(event),
		PagePath:  strings.TrimSpace(path),
	}).FirstOrCreate(&o).Error
	if err != nil {
		return nil, fmt.Errorf("creating goal%w", err)
	}
	return &o, nil
}

func GoalByEvent(ctx context.Context, domain, event string) (*Goal, error) {
	var g Goal
	err := Get(ctx).Model(&Goal{}).Where("domain = ?", domain).
		Where("event_name = ?", event).
		Find(&g).Error
	if err != nil {
		return nil, fmt.Errorf("goal by event=%q%w", event, err)
	}
	return &g, nil
}

func GoalByPage(ctx context.Context, domain, page string) (*Goal, error) {
	var g Goal
	err := Get(ctx).Model(&Goal{}).Where("domain = ?", domain).
		Where("page_path = ?", page).
		Find(&g).Error
	if err != nil {
		return nil, fmt.Errorf("goal by page=%q%w", page, err)
	}
	return &g, nil
}

func Goals(ctx context.Context, domain string) (o []*Goal, err error) {
	err = Get(ctx).Model(&Goal{}).Where("domain = ?", domain).Find(&o).Error
	if err != nil {
		err = fmt.Errorf("list goals%w", err)
	}
	return
}

func DeleteGoal(ctx context.Context, gid, domain string) error {
	id, err := strconv.ParseUint(gid, 10, 64)
	if err != nil {
		return fmt.Errorf("parsing goal id=%q domain=%q%w", gid, domain, err)
	}
	err = Get(ctx).Where("domain = ?", domain).Delete(&Goal{
		Model: schema.Model{ID: id},
	}).Error
	if err != nil {
		return fmt.Errorf("deleting goal id=%q domain=%q%w", gid, domain, err)
	}
	return nil
}

type Invitation = schema.Invitation

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Logger = db.Logger.LogMode(logger.Silent)
	db.SetupJoinTable(&User{}, "Sites", &SiteMembership{})
	db.SetupJoinTable(&Site{}, "Users", &SiteMembership{})
	err = db.AutoMigrate(
		&Goal{},
		&Invitation{},
		&SharedLink{},
		&SiteMembership{},
		&Site{},
		&User{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CloseDB(db *gorm.DB) error {
	x, _ := db.DB()
	return x.Close()
}

type dbKey struct{}

func Set(ctx context.Context, dbs *gorm.DB) context.Context {
	return context.WithValue(ctx, dbKey{}, dbs)
}

func Get(ctx context.Context) *gorm.DB {
	return ctx.Value(dbKey{}).(*gorm.DB)
}

func Exists(ctx context.Context, where func(db *gorm.DB) *gorm.DB) bool {
	return exists(Get(ctx), where)
}

func exists(g *gorm.DB, where func(db *gorm.DB) *gorm.DB) bool {
	db := where(g).Select("1").Limit(1)
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
