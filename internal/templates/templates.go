package templates

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vinceanalytics/vince/internal/config"
	"github.com/vinceanalytics/vince/internal/core"
	"github.com/vinceanalytics/vince/internal/flash"
	"github.com/vinceanalytics/vince/internal/models"
	"github.com/vinceanalytics/vince/internal/query"
	"github.com/vinceanalytics/vince/internal/timeseries"
	"github.com/vinceanalytics/vince/pkg/octicon"
	"github.com/vinceanalytics/vince/pkg/timex"
)

//go:embed layout  plot site stats auth error email user
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
		"Logo":       LogoText,
		"GoalName":   models.GoalName,
		"SafeDomain": models.SafeDomain,
		"HumanDate": func(ts time.Time) string {
			return ts.Format(timex.HumanDate)
		},
		"Periods": func() []timex.Duration {
			return []timex.Duration{
				timex.Today,
				timex.ThisWeek,
				timex.ThisMonth,
				timex.ThisYear,
			}
		},
		"SelectedPeriod": func(a, b timex.Duration) string {
			if a != b {
				return "d-none"
			}
			return ""
		},
		"JSON": func(a any) (string, error) {
			b, err := json.Marshal(a)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
		"PATH": PATH,
	})
}

func PATH(b string, a ...string) string {
	q := make(url.Values)
	for i := 0; i < len(a); i += 2 {
		q.Set(a[i], a[i+1])
	}
	if len(b) > 0 && b[0] != '/' {
		b = "/" + b
	}
	return b + "?" + q.Encode()
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

var PasswordResetEmail = template.Must(
	layout().ParseFS(Files,
		"email/password_reset_email.html",
	),
).Lookup("base_email")

var Activate = template.Must(
	layout().ParseFS(Files,
		"auth/activate.html",
	),
).Lookup("focus")

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

var PasswordResetRequestSuccess = template.Must(
	layout().ParseFS(Files,
		"auth/password_reset_request_success.html",
	),
).Lookup("focus")

var PasswordResetForm = template.Must(
	layout().ParseFS(Files,
		"auth/password_reset_form.html",
	),
).Lookup("focus")

var InviteExistingUser = template.Must(
	layout().ParseFS(Files,
		"email/existing_user_invitation.html",
	),
).Lookup("base_email")

var InviteNewUser = template.Must(
	layout().ParseFS(Files,
		"email/new_user_invitation.html",
	),
).Lookup("base_email")

var Home = template.Must(
	layout().ParseFS(Files,
		"user/home.html",
	),
).Lookup("app")

var SiteHome = template.Must(
	layout().ParseFS(Files,
		"site/home.html",
	),
).Lookup("app")

type Errors struct {
	Status     int
	StatusText string
}

// For our logo the font used is Contrail One face 700-bold italic with size 150

type Context struct {
	Title         string
	Header        Header
	USER          *models.User
	Data          map[string]any
	CSRF          template.HTML
	Captcha       template.HTMLAttr
	Errors        map[string]string
	Form          url.Values
	Code          uint64
	ResetLink     string
	Token         string
	Email         string
	Config        *config.Options
	HasInvitation bool
	HasPin        bool
	Flash         *flash.Flash
	Error         *Errors
	Site          *models.Site
	Goals         []*models.Goal
	IsFIrstSite   bool
	SitesOverview []models.SiteOverView
	HasGoals      bool
	Owner         *models.User
	Recipient     string
	SharedLink    *models.SharedLink
	Stats         *timeseries.Stats
	Now           core.NowFunc
	Invite        *Invite
	Overview      *Overview
}

func (t *Context) ProfileOverview() string {
	return "/" + t.USER.Name
}

func (t *Context) ProfileSites() string {
	q := make(url.Values)
	q.Set("panel", "sites")
	return "/" + t.USER.Name + "?" + q.Encode()
}

type Header struct {
	Context    string
	Mode       string
	ContextRef string
}

type Overview struct {
	Global query.Global
	Sites  []SiteOverView
	Panel  string
}

type SiteOverView struct {
	Site   *models.Site
	Owner  string
	Global query.Global
}

type Invite struct {
	Email string
	ID    uint64
}

func (t *Context) InviteURL() string {
	if t.Invite.ID == 0 {
		return fmt.Sprintf("%s/sites", t.Config.URL)
	}
	return fmt.Sprintf("%s/register/invitation/%d", t.Config.URL, t.Invite.ID)
}

func (t *Context) GreetRecipient() string {
	if t.Recipient == "" {
		return "Hey"
	}
	return "Hey " + strings.Split(t.Recipient, " ")[0]
}

func New(ctx context.Context, f ...func(c *Context)) *Context {
	c := &Context{
		Data:    make(map[string]any),
		CSRF:    getCsrf(ctx),
		Captcha: getCaptcha(ctx),
		USER:    models.GetUser(ctx),
		Config:  config.Get(ctx),
		Flash:   flash.Get(ctx),
		Errors:  make(map[string]string),
		Form:    make(url.Values),
		Now:     core.GetNow(ctx),
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

func (t *Context) Home() string {
	return t.Config.URL + "/"
}

func (t *Context) Snippet() string {
	track := fmt.Sprintf("%s/js/vince.js", t.Config.URL)
	src := fmt.Sprintf("<script defer data-domain=%q src=%q></script>", models.SafeDomain(t.Site), track)
	return src
}

func (t *Context) SharedLinkURL(site *models.Site, link *models.SharedLink) string {
	return models.SharedLinkURL(t.Config.URL, site, link)
}

func Avatar(uid string, size uint, class ...string) template.HTML {
	q := make(url.Values)
	q.Set("u", uid)
	q.Set("s", strconv.Itoa(int(size)))
	u := "/avatar?" + q.Encode()
	return template.HTML(fmt.Sprintf(`<img class=%q width="%d" height="%d" src=%q>`,
		strings.Join(class, " "), size, size, u,
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
