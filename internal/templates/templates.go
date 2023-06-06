package templates

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/flash"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/octicon"
	"github.com/vinceanalytics/vince/pkg/timex"
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
		"GoalName":   models.GoalName,
		"SafeDomain": models.SafeDomain,
		"PeriodLabel": func(ts time.Time) string {
			return ts.Format("Jan 02, 2006")
		},
		"HumanDate": func(ts time.Time) string {
			return ts.Format(timex.HumanDate)
		},
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
).Lookup("focus")

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

var UserSettings = template.Must(
	layout().ParseFS(Files,
		"auth/user_settings.html",
	),
).Lookup("app")

var SiteSettings = template.Must(
	layout().ParseFS(Files,
		"site/settings.html",
	),
).Lookup("app")

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

var SharedLinkForm = template.Must(
	layout().ParseFS(Files,
		"site/new_shared_link.html",
	),
).Lookup("focus")

var EditSharedLinkForm = template.Must(
	layout().ParseFS(Files,
		"site/edit_shared_link.html",
	),
).Lookup("focus")

var PasswordResetRequestForm = template.Must(
	layout().ParseFS(Files,
		"auth/password_reset_request_form.html",
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
	Config        *config.Options
	HasInvitation bool
	HasPin        bool
	Flash         *flash.Flash
	NewSite       *NewSite
	Error         *Errors
	Site          *models.Site
	Goals         []*models.Goal
	IsFIrstSite   bool
	SitesOverview []models.SiteOverView
	EmailReport   bool
	HasGoals      bool
	Owner         *models.User
	Recipient     string
	Key           string
	SharedLink    *models.SharedLink
	Stats         *timeseries.Stats
	Now           core.NowFunc
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
		Now:         core.GetNow(ctx),
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

func (t *Context) Snippet() string {
	track := fmt.Sprintf("%s/js/vince.js", t.Config.Url)
	src := fmt.Sprintf("<script defer data-domain=%q src=%q></script>", models.SafeDomain(t.Site), track)
	return src
}

func (t *Context) SharedLinkURL(site *models.Site, link *models.SharedLink) string {
	return models.SharedLinkURL(t.Config.Url, site, link)
}

func Avatar(uid uint64, size uint, class ...string) template.HTML {
	return template.HTML(fmt.Sprintf(`<img class=%q src="/avatar?u=%d&s=%d">`,
		strings.Join(class, " "), uid, size,
	))
}

func (t *Context) SiteIndex() (o [][]models.SiteOverView) {
	var m []models.SiteOverView
	for _, v := range t.SitesOverview {
		m = append(m, v)
		if len(m) == 3 {
			o = append(o, m)
			m = nil
		}
	}
	if len(m) > 0 {
		o = append(o, m)
	}
	return
}
