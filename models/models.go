package models

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/pkg/log"
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

func UpdateSiteStartDate(ctx context.Context, sid uint64, start time.Time) {
	err := Get(ctx).Model(&Site{}).Where("id = ?", sid).Update("stats_start_date", start).Error
	if err != nil {
		DBE(ctx, err, "failed to update stats_start_date")
	}
}

func EnableWeeklyReport(ctx context.Context, site *Site, usr *User) {
	site.Preload(ctx, "WeeklyReport")
	if site.WeeklyReport != nil {
		// This is a work around for storing arrays in sqlite. We use comma separated
		// list for emails.
		site.WeeklyReport.Recipients += "," + usr.Email
		err := Get(ctx).Save(site.WeeklyReport).Error
		if err != nil {
			DBE(ctx, err, "failed to update weekly update recipients", func(e *zerolog.Event) *zerolog.Event {
				return e.Uint64("sid", site.ID).Uint64("uid", usr.ID)
			})
		}
		return
	}
	err := Get(ctx).Model(site).Association("WeeklyReport").Append(&WeeklyReport{
		Recipients: usr.Email,
	})
	if err != nil {
		DBE(ctx, err, "failed to create weekly report for site", func(e *zerolog.Event) *zerolog.Event {
			return e.Uint64("sid", site.ID).Uint64("uid", usr.ID)
		})
	}
}

func DeleteSite(ctx context.Context, site *Site) {
	err := Get(ctx).Select("SiteMemberships").Delete(site).Error
	if err != nil {
		DBE(ctx, err, "failed to delete site")
	}
}

func (s *Site) SafeDomain() string {
	return url.PathEscape(s.Domain)
}

func (u *Site) Preload(ctx context.Context, preload ...string) {
	db := Get(ctx)
	for _, p := range preload {
		db = db.Preload(p)
	}
	err := db.First(u).Error
	if err != nil {
		DBE(ctx, err, "failed to preload for Site model "+strings.Join(preload, ","))
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

func (s *Site) Owner(ctx context.Context) *User {
	var u User
	err := Get(ctx).Model(&User{}).
		Joins("left join site_memberships on site_memberships.user_id = users.id").
		Where("site_memberships.site_id = ?", s.ID).
		Where("site_memberships.role = ?", "owner").First(&u).Error
	if err != nil {
		DBE(ctx, err, "failed to find site owner")
		return nil
	}
	return &u
}

func (s *Site) HasGoals(ctx context.Context) bool {
	return Exists(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Model(&Goal{}).Where("domain = ?", s.Domain)
	})
}

func SiteByDomain(ctx context.Context, domain string) *Site {
	var s Site
	err := Get(ctx).Model(&Site{}).Where("domain = ?", domain).First(&s).Error
	if err != nil {
		DBE(ctx, err, "failed to find site by domain")
		return nil
	}
	return &s
}

func ChangeSiteVisibility(ctx context.Context, site *Site, public bool) {
	err := Get(ctx).Model(site).Update("public", public).Error
	if err != nil {
		DBE(ctx, err, "failed to change site visibility")
	}
}
func (s *Site) IsMember(ctx context.Context, uid uint64) bool {
	return Role(ctx, uid, s.ID) != ""
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

func UpdateAPIKeyUse(ctx context.Context, aid string) {
	// aid is string because we use value we set in jwt token claims. No need to do
	// extra decoding here.If its invalid value then an error will show up on logs
	err := Get(ctx).Model(&APIKey{}).Where("id = ?", aid).
		Update("used_at", time.Now()).Error
	if err != nil {
		DBE(ctx, err, "failed to update used at time")
	}
}

func APIKeyByID(ctx context.Context, aid string) (a *APIKey) {
	var m APIKey
	err := Get(ctx).Where("id = ?", aid).First(&m).Error
	if err != nil {
		DBE(ctx, err, "failed to get key by id")
		return nil
	}
	return &m
}

func (ak *APIKey) RateLimit() (rate.Limit, int) {
	r := rate.Limit(float64(ak.HourlyAPIRequestLimit) / time.Hour.Seconds())
	return r, 10
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

func DBE(ctx context.Context, err error, msg string, f ...func(*zerolog.Event) *zerolog.Event) {
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, sql.ErrNoRows) {
		return
	}
	if len(f) > 0 {
		f[0](log.Get(ctx).Err(err)).Msg(msg)
	} else {
		log.Get(ctx).Err(err).Msg(msg)
	}
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

func CreateGoal(ctx context.Context, domain, event, path string) {
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
		DBE(ctx, err, "failed to create a new goal")
	}
}

func Goals(ctx context.Context, domain string) (o []*Goal) {
	err := Get(ctx).Model(&Goal{}).Where("domain = ?", domain).Find(&o).Error
	if err != nil {
		DBE(ctx, err, "failed to find goals by domain")
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
		Model: Model{ID: id},
	}).Error
	if err != nil {
		DBE(ctx, err, "failed to delete goal")
		return false
	}
	return true
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
}

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
		&EnterprisePlan{},
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
