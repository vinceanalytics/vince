package templates

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/gernest/vince/config"
	"github.com/gernest/vince/flash"
	"github.com/gernest/vince/models"
	"github.com/gernest/vince/octicon"
)

//go:embed layout  plot site stats auth error email
var Files embed.FS

var Layouts = template.Must(
	base().ParseFS(Files,
		"layout/*.html",
	),
)

var LoginForm = template.Must(
	layout().ParseFS(Files,
		"auth/login_form.html",
	),
).Lookup("focus")

func base() *template.Template {
	return template.New("root").Funcs(template.FuncMap{
		"Icon":       octicon.IconTemplateFunc,
		"Avatar":     Avatar,
		"Logo":       Logo,
		"Calendar":   CalendarEntries,
		"ActiveItem": ActiveItem,
		"GoalName":   models.GoalName,
		"SafeDomain": models.SafeDomain,
		"ThisYear":   thisYearFormat,
	})
}

var RegisterForm = template.Must(
	layout().ParseFS(Files,
		"auth/register_form.html",
	),
).Lookup("focus")

func layout() *template.Template {
	return template.Must(Layouts.Clone())
}

var Error = template.Must(
	template.ParseFS(Files,
		"error/error.html",
	))

var ActivationEmail = template.Must(
	layout().ParseFS(Files,
		"email/activation_code.html",
	),
).Lookup("base_email")

var Activate = template.Must(
	layout().ParseFS(Files,
		"auth/activate.html",
	),
).Lookup("focus")

var Sites = template.Must(
	layout().ParseFS(Files,
		"plot/plot.html",
		"site/index.html",
	),
).Lookup("app")

var SiteNew = template.Must(
	layout().ParseFS(Files,
		"site/new.html",
	),
).Lookup("app")

var AddSnippet = template.Must(
	layout().ParseFS(Files,
		"site/snippet.html",
	),
).Lookup("focus")

var Stats = template.Must(
	layout().ParseFS(Files,
		"stats/stats.html",
	),
).Lookup("app")

var WaitingFirstPageView = template.Must(
	layout().ParseFS(Files,
		"stats/stats.html",
	),
).Lookup("app")

var SiteLocked = template.Must(
	layout().ParseFS(Files,
		"stats/site_locked.html",
	),
).Lookup("app")

var UserSettings = template.Must(
	layout().ParseFS(Files,
		"auth/user_settings.html",
	),
).Lookup("app")

var SiteSettingsGeneral = template.Must(
	layout().ParseFS(Files,
		"site/settings_general.html",
	),
).Lookup("site_settings")

var SiteSettingsGoals = template.Must(
	layout().ParseFS(Files,
		"site/settings_goals.html",
	),
).Lookup("site_settings")

var SiteSettingsPeople = template.Must(
	layout().ParseFS(Files,
		"site/settings_people.html",
	),
).Lookup("site_settings")

var SiteSettingsVisibility = template.Must(
	layout().ParseFS(Files,
		"site/settings_visibility.html",
	),
).Lookup("site_settings")

var SiteSettingsReports = template.Must(
	layout().ParseFS(Files,
		"site/settings_email_reports.html",
	),
).Lookup("site_settings")

var SiteSettingsDanger = template.Must(
	layout().ParseFS(Files,
		"site/settings_danger_zone.html",
	),
).Lookup("site_settings")

var SiteNewGoal = template.Must(
	layout().ParseFS(Files,
		"site/new_goal.html",
	),
).Lookup("focus")

var PasswordForm = template.Must(
	layout().ParseFS(Files,
		"auth/password_form.html",
	),
).Lookup("focus")

var InviteMemberForm = template.Must(
	layout().ParseFS(Files,
		"site/invite_member_form.html",
	),
).Lookup("focus")

type NewSite struct {
	IsFirstSite bool
}

type Errors struct {
	Status     int
	StatusText string
}

// For our logo the font used is Contrail One face 700-bold italic with size 150

type Context struct {
	Title         string
	CurrentUser   *models.User
	Data          map[string]any
	CSRF          template.HTML
	Captcha       template.HTMLAttr
	Errors        map[string]string
	Form          url.Values
	Code          uint64
	Config        *config.Config
	HasInvitation bool
	HasPin        bool
	Flash         *flash.Flash
	NewSite       *NewSite
	Error         *Errors
	Page          string
	Site          *models.Site
	Goals         []*models.Goal
	IsFIrstSite   bool
	SitesOverview []models.SiteOverView
	EmailReport   bool
	HasGoals      bool
	Owner         *models.User
	// Name of the email recipient
	Recipient string
}

func (t *Context) GreetRecipient() string {
	if t.Recipient == "" {
		return "Hey"
	}
	return "Hey " + strings.Split(t.Recipient, " ")[0]
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:        make(map[string]any),
		CSRF:        getCsrf(ctx),
		Captcha:     getCaptcha(ctx),
		CurrentUser: models.GetUser(ctx),
		Config:      config.Get(ctx),
		Flash:       flash.Get(ctx),
		Errors:      make(map[string]string),
		Form:        make(url.Values),
	}
	if len(f) > 0 {
		f[0](c)
	}
	return c
}

type csrfTokenCtxKey struct{}

func getCsrf(ctx context.Context) template.HTML {
	if c := ctx.Value(csrfTokenCtxKey{}); c != nil {
		return c.(template.HTML)
	}
	return template.HTML("")
}

type captchaTokenKey struct{}

func SetCaptcha(ctx context.Context, x template.HTMLAttr) context.Context {
	return context.WithValue(ctx, captchaTokenKey{}, x)
}

func SetCSRF(ctx context.Context, x template.HTML) context.Context {
	return context.WithValue(ctx, csrfTokenCtxKey{}, x)
}

func getCaptcha(ctx context.Context) template.HTMLAttr {
	if c := ctx.Value(captchaTokenKey{}); c != nil {
		return c.(template.HTMLAttr)
	}
	return template.HTMLAttr("")
}

func (t *Context) VinceURL() template.HTML {
	return template.HTML("http://localhost:8080")
}

func Logo(width, height int) template.HTML {
	return template.HTML(fmt.Sprintf(
		`<img alt="Vince Analytics logo" width=%d height=%d src=%q>`,
		width, height, "/image/logo.svg",
	))
}

func (Context) Slogan() string {
	return "Self hosted, single file ,privacy friendly web analytics platform"
}

func (t *Context) Snippet() string {
	track := fmt.Sprintf("https://%s/js/vince.js", t.Config.Url)
	src := fmt.Sprintf("<script defer data-domain=%q src=%q></script>", models.SafeDomain(t.Site), track)
	return src
}

func Avatar(uid uint64, size uint, class ...string) template.HTML {
	return template.HTML(fmt.Sprintf(`<img class=%q src="/avatar?u=%d&s=%d">`,
		strings.Join(class, " "), uid, size,
	))
}

func ActiveItem(ctx *Context, key string) string {
	if ctx.Page == key {
		return "ActionListItem--navActive"
	}
	return ""
}
