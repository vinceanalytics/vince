package models

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pkg/db"
	"github.com/gernest/vince/pkg/log"
	"github.com/gernest/vince/pkg/schema"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
	"gorm.io/driver/sqlite"
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

func UpdateSiteStartDate(ctx context.Context, sid uint64, start time.Time) {
	err := Get(ctx).Model(&Site{}).Where("id = ?", sid).Update("stats_start_date", start).Error
	if err != nil {
		LOG(ctx, err, "failed to update stats_start_date")
	}
}

func EnableWeeklyReport(ctx context.Context, site *Site, usr *User) {
	PreloadSite(ctx, site, "WeeklyReport")
	if site.WeeklyReport != nil {
		// This is a work around for storing arrays in sqlite. We use comma separated
		// list for emails.
		site.WeeklyReport.Recipients += "," + usr.Email
		err := Get(ctx).Save(site.WeeklyReport).Error
		if err != nil {
			LOG(ctx, err, "failed to update weekly update recipients", func(e *zerolog.Event) *zerolog.Event {
				return e.Uint64("sid", site.ID).Uint64("uid", usr.ID)
			})
		}
		return
	}
	err := Get(ctx).Model(site).Association("WeeklyReport").Append(&WeeklyReport{
		Recipients: usr.Email,
	})
	if err != nil {
		LOG(ctx, err, "failed to create weekly report for site", func(e *zerolog.Event) *zerolog.Event {
			return e.Uint64("sid", site.ID).Uint64("uid", usr.ID)
		})
	}
}

func DeleteSite(ctx context.Context, site *Site) {
	err := Get(ctx).Select("SiteMemberships").Delete(site).Error
	if err != nil {
		LOG(ctx, err, "failed to delete site")
	}
}

func SafeDomain(s *Site) string {
	return url.PathEscape(s.Domain)
}

func PreloadSite(ctx context.Context, u *Site, preload ...string) {
	db := Get(ctx)
	for _, p := range preload {
		db = db.Preload(p)
	}
	err := db.First(u).Error
	if err != nil {
		LOG(ctx, err, "failed to preload for Site model "+strings.Join(preload, ","))
	}
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

func SiteOwner(ctx context.Context, sid uint64) *User {
	var u User
	err := Get(ctx).Model(&User{}).
		Joins("left join site_memberships on site_memberships.user_id = users.id").
		Where("site_memberships.site_id = ?", sid).
		Where("site_memberships.role = ?", "owner").First(&u).Error
	if err != nil {
		LOG(ctx, err, "failed to find site owner")
		return nil
	}
	return &u
}

func SiteFor(ctx context.Context, uid uint64, domain string, roles ...string) *Site {
	var site Site
	err := Get(ctx).Model(&Site{}).
		Joins("left join site_memberships on site_memberships.site_id = sites.id").
		Where("site_memberships.user_id = ?", uid).
		Where("sites.domain = ?", domain).
		Where("site_memberships.role IN " + buildRoles(roles...)).First(&site).Error
	if err != nil {
		LOG(ctx, err, "failed to find site owner")
		return nil
	}
	return &site
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

func SiteByDomain(ctx context.Context, domain string) *Site {
	var s Site
	err := Get(ctx).Model(&Site{}).Where("domain = ?", domain).First(&s).Error
	if err != nil {
		LOG(ctx, err, "failed to find site by domain")
		return nil
	}
	return &s
}

func ChangeSiteVisibility(ctx context.Context, site *Site, public bool) {
	err := Get(ctx).Model(site).Update("public", public).Error
	if err != nil {
		LOG(ctx, err, "failed to change site visibility")
	}
}

func UserIsMember(ctx context.Context, uid, sid uint64) bool {
	return Role(ctx, uid, sid) != ""
}

type CreateSiteEmail = schema.CreateSiteEmail

type CheckStatEmail = schema.CheckStatEmail

type EmailVerificationCode = schema.EmailVerificationCode

type SpikeNotification = schema.SpikeNotification

type SiteMembership = schema.SiteMembership

type APIKey = schema.APIKey

func HashAPIKey(ctx context.Context, key string) string {
	h := sha256.New()
	h.Write(config.GetSecuritySecret(ctx))
	h.Write([]byte(key))
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}

func ProcessAPIKey(ctx context.Context, key string) (hash, prefix string) {
	return HashAPIKey(ctx, key), key[:6]
}

func CreateApiKey(ctx context.Context, key, name string, uid uint64) {
	hash, prefix := ProcessAPIKey(ctx, key)
	err := Get(ctx).Create(&APIKey{
		Name:      name,
		UserID:    uid,
		KeyPrefix: prefix,
		KeyHash:   hash,
	}).Error
	if err != nil {
		LOG(ctx, err, "failed to create api token")
	}
}

func VerifyAPIKey(ctx context.Context, key string) *APIKey {
	hash := HashAPIKey(ctx, key)
	var a APIKey
	err := Get(ctx).Model(&APIKey{}).Where("key_hash = ?", hash).First(&a).Error
	if err != nil {
		LOG(ctx, err, "failed to get api key")
		return nil
	}
	return &a
}

func UpdateAPIKeyUse(ctx context.Context, aid uint64) {
	// aid is string because we use value we set in jwt token claims. No need to do
	// extra decoding here.If its invalid value then an error will show up on logs
	err := Get(ctx).Model(&APIKey{}).Where("id = ?", aid).
		Update("used_at", time.Now()).Error
	if err != nil {
		LOG(ctx, err, "failed to update used at time")
	}
}

func APIKeyByID(ctx context.Context, aid uint64) (a *APIKey) {
	var m APIKey
	err := Get(ctx).Where("id = ?", aid).First(&m).Error
	if err != nil {
		LOG(ctx, err, "failed to get key by id")
		return nil
	}
	return &m
}

func APIRateLimit(ak *APIKey) (rate.Limit, int) {
	r := rate.Limit(float64(ak.HourlyAPIRequestLimit) / time.Hour.Seconds())
	return r, 10
}

type IntroEmail = schema.IntroEmail

type FeedbackEmail = schema.FeedbackEmail

type Subscription = schema.Subscription

func LOG(ctx context.Context, err error, msg string, f ...func(*zerolog.Event) *zerolog.Event) {
	db.LOG(ctx, err, msg, f...)
}

type SharedLink = schema.SharedLink

type SentRenewalNotification = schema.SentRenewalNotification

type WeeklyReport = schema.WeeklyReport

type SentWeeklyReport = schema.SentWeeklyReport

type MonthlyReport = schema.MonthlyReport

type SentMonthlyReport = schema.SentMonthlyReport

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

func CreateGoal(ctx context.Context, domain, event, path string) *Goal {
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
		LOG(ctx, err, "failed to create a new goal")
	}
	return &o
}

func GoalByEvent(ctx context.Context, domain, event string) *Goal {
	var g Goal
	err := Get(ctx).Model(&Goal{}).Where("domain = ?", domain).
		Where("event_name = ?", event).
		Find(&g).Error
	if err != nil {
		LOG(ctx, err, "failed to find goal by event_name")
		return nil
	}
	return &g
}

func GoalByPage(ctx context.Context, domain, page string) *Goal {
	var g Goal
	err := Get(ctx).Model(&Goal{}).Where("domain = ?", domain).
		Where("page_path = ?", page).
		Find(&g).Error
	if err != nil {
		LOG(ctx, err, "failed to find goal by page_path")
		return nil
	}
	return &g
}

func Goals(ctx context.Context, domain string) (o []*Goal) {
	err := Get(ctx).Model(&Goal{}).Where("domain = ?", domain).Find(&o).Error
	if err != nil {
		LOG(ctx, err, "failed to find goals by domain")
	}
	return
}

func DeleteGoal(ctx context.Context, gid, domain string) bool {
	id, err := strconv.ParseUint(gid, 10, 64)
	if err != nil {
		log.Get(ctx).Err(err).
			Str("domain", domain).
			Str("id", gid).Msg("failed parsing goal id")
		return false
	}
	err = Get(ctx).Where("domain = ?", domain).Delete(&Goal{
		Model: schema.Model{ID: id},
	}).Error
	if err != nil {
		LOG(ctx, err, "failed to delete goal")
		return false
	}
	return true
}

type Invitation = schema.Invitation

func Database(cfg *config.Config) string {
	return filepath.Join(cfg.DataPath, "vince.db")
}

func Open(path string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Logger = db.Logger.LogMode(logger.Silent)
	db.SetupJoinTable(&User{}, "Sites", &SiteMembership{})
	db.SetupJoinTable(&Site{}, "Users", &SiteMembership{})
	err = db.AutoMigrate(
		&APIKey{},
		&CheckStatEmail{},
		&CreateSiteEmail{},
		&EmailVerificationCode{},
		&FeedbackEmail{},
		&Goal{},
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

func CloseDB(db *gorm.DB) error {
	x, _ := db.DB()
	return x.Close()
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

func Set(ctx context.Context, dbs *gorm.DB) context.Context {
	return db.Set(ctx, dbs)
}

func Get(ctx context.Context) *gorm.DB {
	return db.Get(ctx)
}

func Exists(ctx context.Context, where func(db *gorm.DB) *gorm.DB) bool {
	return db.Exists(ctx, where)
}

func exists(g *gorm.DB, where func(db *gorm.DB) *gorm.DB) bool {
	return db.ExistsDB(g, where)
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
